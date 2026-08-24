package handlers

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ResumeHandler struct{ AgentServiceURL string }

func NewResumeHandler(agentServiceURL string) *ResumeHandler {
	return &ResumeHandler{AgentServiceURL: agentServiceURL}
}

func (h *ResumeHandler) Resume(c *gin.Context) {
	sessionID := c.Param("session_id")
	var req struct {
		Decision string `json:"decision" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	body := fmt.Sprintf(`{"decision": "%s"}`, req.Decision)
	resp, err := http.Post(
		h.AgentServiceURL+"/agent/"+sessionID+"/resume",
		"application/json",
		bytes.NewBufferString(body),
	)
	if err != nil || resp.StatusCode >= 400 {
		c.JSON(http.StatusBadGateway, gin.H{"error": "AGENT_RESUME_FAILED"})
		return
	}
	defer resp.Body.Close()

	c.JSON(http.StatusOK, gin.H{"status": "resumed", "session_id": sessionID, "decision": req.Decision})
}
