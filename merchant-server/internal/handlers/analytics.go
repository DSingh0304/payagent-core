package handlers

import (
	"net/http"

	"github.com/DSingh0304/payagent-core/merchant-server/internal/services"
	"github.com/gin-gonic/gin"
)

type AnalyticsHandler struct {
	AuditSvc *services.AuditService
}

func NewAnalyticsHandler(auditSvc *services.AuditService) *AnalyticsHandler {
	return &AnalyticsHandler{AuditSvc: auditSvc}
}

func (h *AnalyticsHandler) GetStats(c *gin.Context) {
	stats := map[string]interface{}{}

	var totalSessions int
	h.AuditSvc.DB.QueryRow(c.Request.Context(), "SELECT COUNT(DISTINCT session_id) FROM audit_logs").Scan(&totalSessions)
	stats["total_sessions"] = totalSessions

	var totalOrders int
	h.AuditSvc.DB.QueryRow(c.Request.Context(), "SELECT COUNT(*) FROM audit_logs WHERE event_type = 'ORDER_CONFIRMED'").Scan(&totalOrders)
	stats["total_orders"] = totalOrders

	var guardrails int
	h.AuditSvc.DB.QueryRow(c.Request.Context(), "SELECT COUNT(*) FROM audit_logs WHERE event_type = 'GUARDRAIL_TRIGGERED'").Scan(&guardrails)
	stats["guardrails_triggered"] = guardrails

	rows, err := h.AuditSvc.DB.Query(c.Request.Context(), `
		SELECT DATE(created_at) as date, COUNT(*) as count 
		FROM audit_logs 
		GROUP BY date 
		ORDER BY date DESC 
		LIMIT 7
	`)
	if err == nil {
		defer rows.Close()
		var timeline []map[string]interface{}
		for rows.Next() {
			var date string
			var count int
			if err := rows.Scan(&date, &count); err == nil {
				timeline = append(timeline, map[string]interface{}{"date": date[:10], "events": count})
			}
		}
		for i, j := 0, len(timeline)-1; i < j; i, j = i+1, j-1 {
			timeline[i], timeline[j] = timeline[j], timeline[i]
		}
		stats["timeline"] = timeline
	} else {
		stats["timeline"] = []map[string]interface{}{}
	}

	c.JSON(http.StatusOK, stats)
}
