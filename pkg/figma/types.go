package figma

import (
	"time"
)

type Comment struct {
	ID         string     `json:"id"`
	CreatedAt  time.Time  `json:"created_at"`
	ResolvedAt *time.Time `json:"resolved_at"`
	User       struct {
		Handle string `json:"handle"`
	} `json:"user"`
	Message    string `json:"message"`
	ClientMeta struct {
		NodeID string `json:"node_id"`
	} `json:"client_meta"`
	ParentID string `json:"parent_id"`
}

type FileNodes struct {
	Name  string          `json:"name"`
	Nodes map[string]Node `json:"nodes"`
}

type Node struct {
	Document struct {
		Name string `json:"name"`
		ID   string `json:"id"`
	} `json:"document"`
}