package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

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
