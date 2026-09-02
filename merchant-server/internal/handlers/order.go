package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/DSingh0304/payagent-core/merchant-server/internal/models"
	"github.com/DSingh0304/payagent-core/merchant-server/internal/services"
)

type OrderHandler struct {
	DB       *pgxpool.Pool
	AuditSvc *services.AuditService
}

func NewOrderHandler(db *pgxpool.Pool, audit *services.AuditService) *OrderHandler {
	return &OrderHandler{DB: db, AuditSvc: audit}
}

func (h *OrderHandler) Create(c *gin.Context) {
	var req struct {
		SessionID       string      `json:"session_id" binding:"required"`
		RazorpayOrderID string      `json:"razorpay_order_id" binding:"required"`
		AmountPaise     int64       `json:"amount_paise" binding:"required"`
		Items           interface{} `json:"items" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var orderID string
	err := h.DB.QueryRow(c.Request.Context(), `
		INSERT INTO orders (session_id, razorpay_order_id, amount_paise, items, status)
		VALUES ($1, $2, $3, $4, 'pending')
		RETURNING id
	`, req.SessionID, req.RazorpayOrderID, req.AmountPaise, req.Items).Scan(&orderID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order", "details": err.Error()})
		return
	}

	payloadBytes, _ := json.Marshal(map[string]interface{}{
		"razorpay_order_id": req.RazorpayOrderID,
		"id":                orderID,
		"amount_paise":      req.AmountPaise,
	})

	h.AuditSvc.Write(c.Request.Context(), models.AuditLog{
		SessionID: req.SessionID,
		EventType: "ORDER_CREATED",
		Actor:     "system",
		Reasoning: "Order created successfully for Razorpay payment",
		Outcome:   "success",
		Payload:   payloadBytes,
	})

	c.JSON(http.StatusOK, gin.H{"id": orderID, "razorpay_order_id": req.RazorpayOrderID, "status": "pending"})
}

func (h *OrderHandler) Confirm(c *gin.Context) {
	orderID := c.Param("order_id")

	var sessionID string
	err := h.DB.QueryRow(c.Request.Context(), `
		UPDATE orders SET status = 'paid' WHERE id = $1 RETURNING session_id
	`, orderID).Scan(&sessionID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order status", "details": err.Error()})
		return
	}

	h.AuditSvc.Write(c.Request.Context(), models.AuditLog{
		SessionID: sessionID,
		EventType: "ORDER_CONFIRMED",
		Actor:     "user",
		Reasoning: "Payment confirmed by user via Razorpay",
		Outcome:   "success",
	})

	c.JSON(http.StatusOK, gin.H{"status": "paid"})
}
