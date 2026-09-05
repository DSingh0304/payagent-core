package handlers

import (
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/DSingh0304/payagent-core/merchant-server/internal/models"
)

type StreamHandler struct {
	DB    *pgxpool.Pool
	Redis *redis.Client
}

func NewStreamHandler(db *pgxpool.Pool, rdb *redis.Client) *StreamHandler {
	return &StreamHandler{DB: db, Redis: rdb}
}

func (h *StreamHandler) Stream(c *gin.Context) {
	sessionID := c.Param("session_id")

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// First, stream historical events from PostgreSQL to instantly hydrate the UI on reload.
	rows, err := h.DB.Query(c.Request.Context(), `
		SELECT id, session_id, event_type, actor, payload, reasoning, outcome, created_at
		FROM audit_logs WHERE session_id = $1 ORDER BY id ASC
	`, sessionID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var l models.AuditLog
			rows.Scan(&l.ID, &l.SessionID, &l.EventType, &l.Actor, &l.Payload, &l.Reasoning, &l.Outcome, &l.CreatedAt)
			data, _ := json.Marshal(l)
			fmt.Fprintf(c.Writer, "event: audit_log\ndata: %s\n\n", data)
			c.Writer.Flush()
		}
	}

	// Subscribe to Redis for real-time live events emitted by the Python LangGraph agent.
	sub := h.Redis.Subscribe(c.Request.Context(), "audit:"+sessionID)
	defer sub.Close()
	ch := sub.Channel()

	for {
		select {
		case msg := <-ch:
			fmt.Fprintf(c.Writer, "event: audit_log\ndata: %s\n\n", msg.Payload)
			c.Writer.Flush()
		case <-c.Request.Context().Done():
			return
		}
	}
}
