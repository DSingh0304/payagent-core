package models

import (
	"encoding/json"
	"time"
)

type AuditLog struct {
	ID        int64           `json:"id"`
	SessionID string          `json:"session_id"`
	EventType string          `json:"event_type"`
	Actor     string          `json:"actor"`
	Payload   json.RawMessage `json:"payload,omitempty"`
	Reasoning string          `json:"reasoning"`
	Outcome   string          `json:"outcome"`
	CreatedAt time.Time       `json:"created_at"`
}
