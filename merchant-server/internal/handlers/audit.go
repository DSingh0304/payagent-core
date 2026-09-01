package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/DSingh0304/payagent-core/merchant-server/internal/models"
	"github.com/DSingh0304/payagent-core/merchant-server/internal/services"
)

type AuditHandler struct{ AuditSvc *services.AuditService }

func NewAuditHandler(audit *services.AuditService) *AuditHandler {
	return &AuditHandler{AuditSvc: audit}
}

func (h *AuditHandler) GetBySession(c *gin.Context) {
	logs, err := h.AuditSvc.GetBySession(c.Request.Context(), c.Param("session_id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"logs": logs, "total": len(logs)})
}

func (h *AuditHandler) WriteAudit(c *gin.Context) {
	var req struct {
		SessionID string `json:"session_id" binding:"required"`
		EventType string `json:"event_type" binding:"required"`
		Actor     string `json:"actor" binding:"required"`
		Reasoning string `json:"reasoning"`
		Outcome   string `json:"outcome"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.AuditSvc.Write(c.Request.Context(), models.AuditLog{
		SessionID: req.SessionID,
		EventType: req.EventType,
		Actor:     req.Actor,
		Reasoning: req.Reasoning,
		Outcome:   req.Outcome,
	})
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *AuditHandler) ListSessions(c *gin.Context) {
	rows, err := h.AuditSvc.DB.Query(c.Request.Context(), `
		SELECT session_id,
			   COALESCE(
			       MIN(reasoning) FILTER (WHERE event_type = 'USER_GOAL'),
			       MIN(reasoning) FILTER (WHERE event_type = 'CATALOG_SEARCH')
			   ) as goal,
			   MAX(event_type) as last_event,
			   COUNT(*) as event_count,
			   MIN(created_at) as started_at,
			   MAX(created_at) as last_activity
		FROM audit_logs
		GROUP BY session_id
		ORDER BY MAX(created_at) DESC
		LIMIT 50
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var sessions []map[string]interface{}
	for rows.Next() {
		var sessionID string
		var goal *string
		var lastEvent string
		var eventCount int
		var startedAt, lastActivity time.Time
		
		if err := rows.Scan(&sessionID, &goal, &lastEvent, &eventCount, &startedAt, &lastActivity); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		sessions = append(sessions, map[string]interface{}{
			"session_id":    sessionID,
			"goal":          goal,
			"last_event":    lastEvent,
			"event_count":   eventCount,
			"started_at":    startedAt,
			"last_activity": lastActivity,
		})
	}

	c.JSON(http.StatusOK, gin.H{"sessions": sessions})
}
