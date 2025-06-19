package figma

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func GetComments(fileKey, token string) ([]Comment, error) {
	url := fmt.Sprintf("https://api.figma.com/v1/files/%s/comments", fileKey)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-FIGMA-TOKEN", token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	var response struct {
		Comments []Comment `json:"comments"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response.Comments, nil
}

func GetFileNodes(fileKey, token string, nodeIDs []string) (*FileNodes, error) {
	params := url.Values{}
	params.Add("ids", strings.Join(nodeIDs, ","))

	url := fmt.Sprintf("https://api.figma.com/v1/files/%s/nodes?%s&depth=1",
		fileKey, params.Encode())

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-FIGMA-TOKEN", token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	var nodesResponse FileNodes
	if err := json.NewDecoder(resp.Body).Decode(&nodesResponse); err != nil {
		return nil, err
	}

	return &nodesResponse, nil
}

func FilterParentComments(comments []Comment) ([]Comment, []string) {
	var parentComments []Comment
	nodeIDMap := make(map[string]bool)

	for _, comment := range comments {
		if comment.ParentID == "" && comment.ClientMeta.NodeID != "" {
			parentComments = append(parentComments, comment)
			nodeIDMap[comment.ClientMeta.NodeID] = true
		}
	}

	nodeIDs := make([]string, 0, len(nodeIDMap))
	for id := range nodeIDMap {
		nodeIDs = append(nodeIDs, id)
	}

	return parentComments, nodeIDs
}