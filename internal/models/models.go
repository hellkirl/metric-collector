package models

type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

type AgentMetrics struct {
	ID    string `json:"id"`
	MType string `json:"type"`
	Delta any    `json:"delta,omitempty"`
	Value any    `json:"value,omitempty"`
}
