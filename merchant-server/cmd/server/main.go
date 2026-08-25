package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/DSingh0304/payagent-core/merchant-server/internal/config"
	appdb "github.com/DSingh0304/payagent-core/merchant-server/internal/db"
	"github.com/DSingh0304/payagent-core/merchant-server/internal/handlers"
	"github.com/DSingh0304/payagent-core/merchant-server/internal/middleware"
	"github.com/DSingh0304/payagent-core/merchant-server/internal/services"
)

func main() {
	cfg := config.Load()

	db := appdb.NewPostgresPool(cfg.PostgresDSN)
	rdb := appdb.NewRedisClient(cfg.RedisURL)

	auditSvc := services.NewAuditService(db, rdb)
	catalogSvc := services.NewCatalogService(db)
	cartSvc := services.NewCartService(rdb, catalogSvc)

	catalogH := handlers.NewCatalogHandler(catalogSvc, auditSvc)
	cartH := handlers.NewCartHandler(cartSvc, auditSvc)
	auditH := handlers.NewAuditHandler(auditSvc)
	streamH := handlers.NewStreamHandler(db, rdb)
	resumeH := handlers.NewResumeHandler(cfg.AgentServiceURL)

	r := gin.Default()

	// Enable CORS for dashboard UI communication
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Content-Type, X-API-Key")
		c.Header("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	r.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	r.GET("/audit/:session_id", auditH.GetBySession)
	r.GET("/stream/:session_id", streamH.Stream)
	r.POST("/agent/:session_id/resume", resumeH.Resume)

	api := r.Group("/api/v1", middleware.AgentAuth(cfg.MerchantAPIKey))
	{
		api.GET("/catalog/search", catalogH.Search)
		api.GET("/cart/:session_id", cartH.Get)
		api.POST("/cart/:session_id/add", cartH.AddItem)
		api.DELETE("/cart/:session_id/remove/:product_id", cartH.RemoveItem)
	}

	log.Printf("Merchant server starting on port :%s", cfg.Port)
	r.Run(":" + cfg.Port)
}
