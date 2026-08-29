package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/DSingh0304/payagent-core/merchant-server/internal/models"
	"github.com/DSingh0304/payagent-core/merchant-server/internal/services"
)

type CatalogHandler struct {
	CatalogSvc *services.CatalogService
	AuditSvc   *services.AuditService
}

func NewCatalogHandler(cat *services.CatalogService, audit *services.AuditService) *CatalogHandler {
	return &CatalogHandler{CatalogSvc: cat, AuditSvc: audit}
}

func (h *CatalogHandler) Search(c *gin.Context) {
	query := c.Query("q")
	category := c.Query("category")
	maxPriceINR, _ := strconv.ParseFloat(c.Query("max_price_inr"), 64)
	
	// Convert INR to paise for accurate monetary comparisons
	maxPricePaise := int64(maxPriceINR * 100)

	products, err := h.CatalogSvc.Search(c.Request.Context(), query, category, maxPricePaise)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if products == nil {
		products = []models.Product{}
	}
	
	h.AuditSvc.Write(c.Request.Context(), models.AuditLog{
		SessionID: c.GetHeader("X-Session-ID"),
		EventType: "CATALOG_SEARCH",
		Actor:     "agent",
		Reasoning: "Searched catalog for: " + query,
		Outcome:   "success",
	})

	c.JSON(http.StatusOK, gin.H{"products": products, "total": len(products)})
}
