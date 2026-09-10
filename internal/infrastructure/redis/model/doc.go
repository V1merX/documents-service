package model

import (
	"encoding/json"
	"time"
	"uuid"
)

type Document struct {
	ID      uuid.UUID       `json:"id"`
	Owner   string          `json:"owner"`
	Name    string          `json:"name"`
	Mime    string          `json:"mime"`
	File    bool            `json:"file"`
	Public  bool            `json:"public"`
	JSON    json.RawMessage `json:"json,omitempty"`
	Content []byte          `json:"content,omitempty"`
	Created time.Time       `json:"created"`
	Grant   []string        `json:"grant"`
}
