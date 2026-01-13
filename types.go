package pi

import "encoding/json"

type RpcCommand map[string]any

type Response struct {
	ID      string          `json:"id,omitempty"`
	Type    string          `json:"type"`
	Command string          `json:"command,omitempty"`
	Success bool            `json:"success"`
	Error   string          `json:"error,omitempty"`
	Data    json.RawMessage `json:"data,omitempty"`
}

type Event struct {
	Type string          `json:"type"`
	Raw  json.RawMessage `json:"-"`
}

type ImageSource struct {
	Type      string `json:"type"`
	MediaType string `json:"mediaType,omitempty"`
	Data      string `json:"data,omitempty"`
	URL       string `json:"url,omitempty"`
}

type ImageContent struct {
	Type   string      `json:"type"`
	Source ImageSource `json:"source"`
}

type Usage struct {
	Input      int   `json:"input"`
	Output     int   `json:"output"`
	CacheRead  int   `json:"cacheRead"`
	CacheWrite int   `json:"cacheWrite"`
	Cost       *Cost `json:"cost,omitempty"`
}

type Cost struct {
	Input      float64 `json:"input"`
	Output     float64 `json:"output"`
	CacheRead  float64 `json:"cacheRead"`
	CacheWrite float64 `json:"cacheWrite"`
	Total      float64 `json:"total"`
}

type RunResult struct {
	Text  string
	Usage *Usage
}
