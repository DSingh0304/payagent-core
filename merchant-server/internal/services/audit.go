package services

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/DSingh0304/payagent-core/merchant-server/internal/models"
)

type AuditService struct {
	DB    *pgxpool.Pool
	Redis *redis.Client
}

func NewAuditService(db *pgxpool.Pool, rdb *redis.Client) *AuditService {
	return &AuditService{DB: db, Redis: rdb}
}

func (s *AuditService) Write(ctx context.Context, entry models.AuditLog) error {
	entry.CreatedAt = time.Now()

	_, err := s.DB.Exec(ctx, `
		INSERT INTO audit_logs (session_id, event_type, actor, payload, reasoning, outcome, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, entry.SessionID, entry.EventType, entry.Actor,
		entry.Payload, entry.Reasoning, entry.Outcome, entry.CreatedAt)
	if err != nil {
		log.Printf("Failed to write audit log: %v", err)
		return err
	}

	data, _ := json.Marshal(entry)
	s.Redis.Publish(ctx, "audit:"+entry.SessionID, string(data))
	return nil
}

func (s *AuditService) GetBySession(ctx context.Context, sessionID string) ([]models.AuditLog, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT id, session_id, event_type, actor, payload, reasoning, outcome, created_at
		FROM audit_logs WHERE session_id = $1 ORDER BY id ASC
	`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.AuditLog
	for rows.Next() {
		var l models.AuditLog
		rows.Scan(&l.ID, &l.SessionID, &l.EventType, &l.Actor, &l.Payload, &l.Reasoning, &l.Outcome, &l.CreatedAt)
		logs = append(logs, l)
	}
	return logs, nil
}
