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

type FigmaComment struct {
	ID         string     `json:"id"`
	CreatedAt  string     `json:"created_at"`
	ResolvedAt *time.Time `json:"resolved_at"`
	User       FigmaUser  `json:"user"`
	Message    string     `json:"message"`
	ClientMeta struct {
		X float64 `json:"x"`
		Y float64 `json:"y"`
	} `json:"client_meta"`
	ParentID string `json:"parent_id"`
	Status   string `json:"status"` // "open" или "resolved"
}

type FigmaUser struct {
	Handle string `json:"handle"`
	ID     string `json:"id"`
}

type Config struct {
	FileKey string `json:"file_key"`
	Output  string `json:"output"`
	Token   string `json:"token"`
}

func main() {
	config, err := loadOrCreateConfig()
	if err != nil {
		log.Fatalf("Ошибка конфигурации: %v", err)
	}

	comments, err := getFigmaComments(config.FileKey, config.Token)
	if err != nil {
		log.Fatalf("Ошибка получения комментариев: %v", err)
	}

	if err := exportToCSV(comments, config.Output); err != nil {
		log.Fatalf("Ошибка экспорта: %v", err)
	}

	fmt.Printf("Экспортировано %d комментариев в %s\n", len(comments), config.Output)
	fmt.Println("Нажмите Enter чтобы выйти...")
	fmt.Scanln()
}

func loadOrCreateConfig() (*Config, error) {
	const configFile = "config.json"

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		defaultConfig := &Config{
			FileKey: "d0kXYmx5RkCkL1lExeAHFX",
			Output:  "comments.csv",
			Token:   "figd_CDl4iC5gX7_Yd4G-h8HWpsgzgsmavTf6kinTEh0c",
		}

		data, err := json.MarshalIndent(defaultConfig, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("ошибка создания конфига: %v", err)
		}

		if err := os.WriteFile(configFile, data, 0644); err != nil {
			return nil, fmt.Errorf("ошибка записи конфига: %v", err)
		}

		return nil, fmt.Errorf("файл конфигурации создан. Заполните config.json и перезапустите программу")
	}

	file, err := os.Open(configFile)
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия конфига: %v", err)
	}
	defer file.Close()

	var config Config
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return nil, fmt.Errorf("ошибка чтения конфига: %v", err)
	}

	if config.FileKey == "" || config.Token == "" {
		return nil, fmt.Errorf("заполните все поля в config.json")
	}

	return &config, nil
}

func getFigmaComments(fileKey, token string) ([]FigmaComment, error) {
	url := fmt.Sprintf("https://api.figma.com/v1/files/%s/comments", fileKey)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API ошибка: %s", resp.Status)
	}

	var result struct {
		Comments []FigmaComment `json:"comments"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var rootComments []FigmaComment
	for _, comment := range result.Comments {
		if comment.ParentID == "" {
			rootComments = append(rootComments, comment)
		}
	}

	return rootComments, nil
}

func exportToCSV(comments []FigmaComment, path string) error {
    file, err := os.Create(path)
    if err != nil {
        return err
    }
    defer file.Close()

    writer := csv.NewWriter(file)
    defer writer.Flush()

    // Добавляем новые заголовки
    headers := []string{
        "ID",
        "Дата создания",
        "Статус",
        "Дата разрешения",
        "Пользователь",
        "ID пользователя",
        "Сообщение",
        "X",
        "Y",
    }

    if err := writer.Write(headers); err != nil {
        return err
    }

    for _, comment := range comments {
        resolvedAt := ""
        if comment.ResolvedAt != nil {
            resolvedAt = comment.ResolvedAt.Format(time.RFC3339)
        }

        record := []string{
            comment.ID,
            comment.CreatedAt,
            comment.Status,
            resolvedAt,
            comment.User.Handle,
            comment.User.ID,
            comment.Message,
            fmt.Sprintf("%.2f", comment.ClientMeta.X),
            fmt.Sprintf("%.2f", comment.ClientMeta.Y),
        }
        if err := writer.Write(record); err != nil {
            return err
        }
    }

    return nil
}
