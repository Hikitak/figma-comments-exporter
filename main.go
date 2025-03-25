package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type FigmaCommentResponse struct {
	Comments []FigmaComment `json:"comments"`
}

type FigmaComment struct {
	ID         string     `json:"id"`
	CreatedAt  time.Time  `json:"created_at"`
	ResolvedAt *time.Time `json:"resolved_at"`
	User       struct {
		Handle string `json:"handle"`
		ID     string `json:"id"`
	} `json:"user"`
	Message    string `json:"message"`
	ClientMeta struct {
		NodeID string `json:"node_id"`
	} `json:"client_meta"`
	ParentID string `json:"parent_id"`
}

type FigmaNodesResponse struct {
	Name  string          `json:"name"`
	Nodes map[string]Node `json:"nodes"`
}

type Node struct {
	Document struct {
		Name string `json:"name"`
		ID   string `json:"id"`
	} `json:"document"`
}

type Config struct {
	FileKeys []string `json:"file_keys"`
	Output   string   `json:"output"`
	Token    string   `json:"token"`
}

func main() {
	config := loadConfig()
	file := createOutputFile(config.Output)
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()
	writeBOM(file)
	writer.Comma = ';'

	writeHeaders(writer)

	for _, fileKey := range config.FileKeys {
		processFile(fileKey, config.Token, writer)
	}

	fmt.Printf("\nЭкспорт завершен. Файл: %s\n", config.Output)
	fmt.Println("Нажмите Enter для выхода...")
	fmt.Scanln()
}

func loadConfig() *Config {
	data, err := os.ReadFile("config.json")
	if err != nil {
		log.Fatal("Ошибка чтения config.json: ", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		log.Fatal("Ошибка парсинга config.json: ", err)
	}

	if len(config.FileKeys) == 0 || config.Token == "" {
		log.Fatal("Некорректная конфигурация")
	}

	return &config
}

func processFile(fileKey, token string, writer *csv.Writer) {
	fmt.Printf("Обработка файла %s... ", fileKey)

	// 1. Получение комментариев
	comments, err := getComments(fileKey, token)
	if err != nil {
		log.Printf("\nОшибка получения комментариев: %v", err)
		return
	}

	// 2. Фильтрация и сбор node_id
	parentComments, nodeIDs := filterParentComments(comments)
	if len(nodeIDs) == 0 {
		fmt.Println("Нет комментариев")
		return
	}

	// 3. Получение информации о нодах
	nodesResponse, err := getNodes(fileKey, token, nodeIDs)
	if err != nil {
		log.Printf("\nОшибка получения нод: %v", err)
		return
	}

	// 4. Формирование и запись данных
	for _, comment := range parentComments {
		nodeID := comment.ClientMeta.NodeID
		if node, exists := nodesResponse.Nodes[nodeID]; exists {
			writeRecord(writer, fileKey, nodesResponse.Name, node, comment)
		}
	}

	fmt.Println("Успешно")
}

func getComments(fileKey, token string) ([]FigmaComment, error) {
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

	var commentsResponse FigmaCommentResponse
	if err := json.NewDecoder(resp.Body).Decode(&commentsResponse); err != nil {
		return nil, err
	}

	return commentsResponse.Comments, nil
}

func filterParentComments(comments []FigmaComment) ([]FigmaComment, []string) {
	var parentComments []FigmaComment
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

func getNodes(fileKey, token string, nodeIDs []string) (*FigmaNodesResponse, error) {
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

	var nodesResponse FigmaNodesResponse
	if err := json.NewDecoder(resp.Body).Decode(&nodesResponse); err != nil {
		return nil, err
	}

	return &nodesResponse, nil
}

func writeRecord(writer *csv.Writer, fileId, fileName string, node Node, comment FigmaComment) {
	resolvedAtToDisplay := ""
	if comment.ResolvedAt != nil {
		resolvedAtToDisplay = comment.ResolvedAt.Format(time.RFC3339)
	} else {
		resolvedAtToDisplay = ""
	}

	record := []string{
		fileName,
		fileId,
		node.Document.Name,
		node.Document.ID,
		comment.Message,
		comment.User.Handle,
		comment.CreatedAt.Format(time.RFC3339),
		getCommentStatus(comment.ResolvedAt),
		resolvedAtToDisplay,
		fmt.Sprintf("https://www.figma.com/design/%s?node-id=%s#%s",
			fileId, strings.Replace(comment.ClientMeta.NodeID, ":", "-", 1), comment.ID),
	}
	writer.Write(record)
}

func getCommentStatus(resolvedAt *time.Time) string {
	if resolvedAt != nil {
		return "resolved"
	}
	return "open"
}

func createOutputFile(filename string) *os.File {
	file, err := os.Create(filename)
	if err != nil {
		log.Fatal("Ошибка создания файла: ", err)
	}
	return file
}

func writeBOM(file *os.File) {
	file.Write([]byte{0xEF, 0xBB, 0xBF})
}

func writeHeaders(writer *csv.Writer) {
	headers := []string{
		"Файл",
		"ID файла",
		"Фрейм",
		"ID фрейма",
		"Комментарий",
		"Автор",
		"Дата создания",
		"Статус",
		"Дата закрытия",
		"Ссылка",
	}
	writer.Write(headers)
}
