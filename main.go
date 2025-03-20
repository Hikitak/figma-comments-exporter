package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type FigmaFileResponse struct {
	Name     string `json:"name"`
	Document struct {
		Children []Node `json:"children"`
	} `json:"document"`
}

type Node struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Children []Node `json:"children,omitempty"`
	Type     string `json:"type"`
}

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
		NodeID string  `json:"node_id"`
		X      float64 `json:"x"`
		Y      float64 `json:"y"`
	} `json:"client_meta"`
	ParentID string `json:"parent_id"`
}

// Конфигурация
type Config struct {
	Files  []FileConfig `json:"files"`
	Output string       `json:"output"`
	Token  string       `json:"token"`
}

type FileConfig struct {
	Key  string `json:"key"`
	Name string `json:"name,omitempty"`
}

func main() {
	config := loadConfig()
	prepareOutput(config.Output)
	processFiles(config)
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

	if len(config.Files) == 0 || config.Token == "" {
		log.Fatal("Некорректная конфигурация")
	}

	return &config
}

func processFiles(config *Config) {
	file, err := os.Create(config.Output)
	if err != nil {
		log.Fatal("Ошибка создания файла: ", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()
	file.Write([]byte{0xEF, 0xBB, 0xBF})
	writer.Comma = ';'

	writeHeaders(writer)

	for _, f := range config.Files {
		processSingleFile(f, config.Token, writer)
	}

	fmt.Printf("\nЭкспортировано в: %s\n", config.Output)
	fmt.Println("Нажмите Enter для выхода...")
	fmt.Scanln()
}

func processSingleFile(file FileConfig, token string, writer *csv.Writer) {
	fmt.Printf("Обработка файла: %s... ", file.Key)

	// 1. Получение комментариев
	comments, nodesIds, err := getComments(file.Key, token)
	if err != nil {
		log.Printf("\nОшибка получения комментариев: %v", err)
		return
	}

	// 2. Получение структуры файла
	fileStructure, err := getFileStructure(nodesIds, file.Key, token)
	if err != nil {
		log.Printf("\nОшибка получения структуры: %v", err)
		return
	}

	// 3. Группировка комментариев по node_id
	commentMap := make(map[string][]FigmaComment)
	for _, c := range comments {
		if c.ParentID == "" && c.ClientMeta.NodeID != "" {
			commentMap[c.ClientMeta.NodeID] = append(commentMap[c.ClientMeta.NodeID], c)
		}
	}

	// 4. Обработка узлов
	fileName := file.Name
	if fileName == "" {
		fileName = fileStructure.Name
	}
	processNodeRecursive(fileStructure.Document.Children, fileName, file.Key, commentMap, writer)

	fmt.Println("Готово")
}

func getFileStructure(nodesIds, fileKey, token string) (*FigmaFileResponse, error) {
	url := fmt.Sprintf("https://api.figma.com/v1/files/%s/nodes?id=%s", fileKey, nodesIds)
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

	var fileResponse FigmaFileResponse
	if err := json.NewDecoder(resp.Body).Decode(&fileResponse); err != nil {
		return nil, err
	}

	return &fileResponse, nil
}

func getComments(fileKey, token string) ([]FigmaComment, string, error) {
	url := fmt.Sprintf("https://api.figma.com/v1/files/%s/comments", fileKey)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-FIGMA-TOKEN", token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	var commentsResponse FigmaCommentResponse
	if err := json.NewDecoder(resp.Body).Decode(&commentsResponse); err != nil {
		return nil, nil, err
	}

	nodesIds := ""
	for _, c := range commentsResponse.Comments {
		if c.ClientMeta.NodeID != nil {
			nodesIds += c.ClientMeta.NodeID
		}
	}

	return commentsResponse.Comments, nodesIds, nil
}

func processNodeRecursive(nodes []Node, fileName, fileKey string, comments map[string][]FigmaComment, writer *csv.Writer) {
	for _, node := range nodes {
		// Обработка текущего узла
		if comments, exists := comments[node.ID]; exists {
			for _, comment := range comments {
				record := []string{
					fileName,
					node.Name,
					node.ID,
					comment.Message,
					comment.User.Handle,
					comment.CreatedAt.Format(time.RFC3339),
					getCommentStatus(comment.ResolvedAt),
					fmt.Sprintf("%.2f", comment.ClientMeta.X),
					fmt.Sprintf("%.2f", comment.ClientMeta.Y),
					fmt.Sprintf("https://www.figma.com/file/%s/?node-id=%s#%s", fileKey, node.ID, comment.ID),
				}
				writer.Write(record)
			}
		}

		// Рекурсивная обработка дочерних узлов
		if len(node.Children) > 0 {
			processNodeRecursive(node.Children, fileName, fileKey, comments, writer)
		}
	}
}

func getCommentStatus(resolvedAt *time.Time) string {
	if resolvedAt != nil {
		return "resolved"
	}
	return "open"
}

func writeHeaders(writer *csv.Writer) {
	headers := []string{
		"Файл",
		"Нода",
		"ID ноды",
		"Комментарий",
		"Автор",
		"Дата создания",
		"Статус",
		"X",
		"Y",
		"Ссылка",
	}
	writer.Write(headers)
}

func prepareOutput(filename string) {
	if _, err := os.Stat(filename); err == nil {
		os.Remove(filename)
	}
}
