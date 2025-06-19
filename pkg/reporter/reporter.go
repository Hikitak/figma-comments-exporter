package reporter

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/tealeg/xlsx"
	"github.com/Hikitak/figma-comment-reporter/pkg/config"
	"github.com/Hikitak/figma-comment-reporter/pkg/figma"
)

type Reporter struct {
	Token    string
	FileKeys []string
	Fields   []config.ReportField
}

func New(token string, fileKeys []string, fields []config.ReportField) *Reporter {
	return &Reporter{
		Token:    token,
		FileKeys: fileKeys,
		Fields:   fields,
	}
}

func (r *Reporter) Generate() ([]byte, error) {
	file := xlsx.NewFile()
	sheet, err := file.AddSheet("Comments")
	if err != nil {
		return nil, err
	}

	headerRow := sheet.AddRow()
	for _, field := range r.Fields {
		cell := headerRow.AddCell()
		cell.Value = field.Display
	}

	for _, fileKey := range r.FileKeys {
		comments, err := figma.GetComments(fileKey, r.Token)
		if err != nil {
			log.Printf("Error getting comments for file %s: %v", fileKey, err)
			continue
		}

		parentComments, nodeIDs := figma.FilterParentComments(comments)
		if len(nodeIDs) == 0 {
			continue
		}

		nodesResponse, err := figma.GetFileNodes(fileKey, r.Token, nodeIDs)
		if err != nil {
			log.Printf("Error getting nodes for file %s: %v", fileKey, err)
			continue
		}

		for _, comment := range parentComments {
			nodeID := comment.ClientMeta.NodeID
			if node, exists := nodesResponse.Nodes[nodeID]; exists {
				row := sheet.AddRow()
				for _, field := range r.Fields {
					value := r.getFieldValue(comment, node, fileKey, nodesResponse.Name, field)
					cell := row.AddCell()
					cell.Value = value
				}
			}
		}
	}

	var buf bytes.Buffer
	if err := file.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (r *Reporter) getFieldValue(comment figma.Comment, node figma.Node, fileID, fileName string, field config.ReportField) string {
	switch field.Name {
	case "file_name":
		return fileName
	case "file_id":
		return fileID
	case "node_name":
		return node.Document.Name
	case "node_id":
		return node.Document.ID
	case "message":
		return comment.Message
	case "author":
		return comment.User.Handle
	case "created_at":
		return formatTime(comment.CreatedAt, field.Format)
	case "status":
		if comment.ResolvedAt != nil {
			return "resolved"
		}
		return "open"
	case "resolved_at":
		if comment.ResolvedAt != nil {
			return formatTime(*comment.ResolvedAt, field.Format)
		}
		return ""
	case "link":
		return fmt.Sprintf("https://www.figma.com/design/%s?node-id=%s#%s",
			fileID, strings.Replace(comment.ClientMeta.NodeID, ":", "-", 1), comment.ID)
	default:
		return ""
	}
}

func formatTime(t time.Time, format string) string {
	if format == "" {
		return t.Format(time.RFC3339)
	}
	return t.Format(format)
}