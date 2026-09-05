package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/DSingh0304/payagent-core/merchant-server/internal/models"
	"github.com/DSingh0304/payagent-core/merchant-server/internal/services"
)

type CartHandler struct {
	CartSvc  *services.CartService
	AuditSvc *services.AuditService
}

func NewCartHandler(cart *services.CartService, audit *services.AuditService) *CartHandler {
	return &CartHandler{CartSvc: cart, AuditSvc: audit}
}

func (h *CartHandler) Get(c *gin.Context) {
	cart, err := h.CartSvc.Get(c.Request.Context(), c.Param("session_id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cart)
}

func (h *CartHandler) AddItem(c *gin.Context) {
	sessionID := c.Param("session_id")
	var req struct {
		ProductID string `json:"product_id" binding:"required"`
		Quantity  int    `json:"quantity" binding:"required,min=1"`
		Reasoning string `json:"agent_reasoning"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cart, err := h.CartSvc.AddItem(c.Request.Context(), sessionID, req.ProductID, req.Quantity, req.Reasoning)
	if err != nil {
		// Differentiate inventory issues (like out-of-stock) for the Python agent.
		// Returning a 409 Conflict allows the agent to gracefully handle the error and suggest alternatives.
		if err.Error() == "STOCK_UNAVAILABLE" {
			c.JSON(http.StatusConflict, gin.H{"error": "STOCK_UNAVAILABLE", "message": "Product out of stock"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.AuditSvc.Write(c.Request.Context(), models.AuditLog{
		SessionID: sessionID,
		EventType: "CART_ADD",
		Actor:     "agent",
		Reasoning: req.Reasoning,
		Outcome:   "success",
	})

	c.JSON(http.StatusOK, gin.H{"cart": cart})
}

func (h *CartHandler) RemoveItem(c *gin.Context) {
	sessionID := c.Param("session_id")
	productID := c.Param("product_id")

	qtyStr := c.Query("qty")
	qty := 0
	if qtyStr != "" {
		fmt.Sscanf(qtyStr, "%d", &qty)
	}

	cart, err := h.CartSvc.RemoveItem(c.Request.Context(), sessionID, productID, qty)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.AuditSvc.Write(c.Request.Context(), models.AuditLog{
		SessionID: sessionID, EventType: "CART_REMOVE", Actor: "agent", Outcome: "success",
	})
	c.JSON(http.StatusOK, gin.H{"cart": cart})
}
