package router

import (
	"github.com/gin-gonic/gin"
	"github.com/uptimebeats/heartbeat-api/internal/api/handlers"
	"github.com/uptimebeats/heartbeat-api/internal/api/middleware"
	"github.com/uptimebeats/heartbeat-api/internal/repository"
)

func SetupRouter(repo *repository.HeartbeatRepository) *gin.Engine {
	r := gin.Default()

	handler := handlers.NewHeartbeatHandler(repo)

	// Root endpoint
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"service":  "Heartbeat Monitoring API",
			"provider": "UptimeBeats.com",
			"register": "https://uptimebeats.com",
			"example":  "http://heartbeat.uptimebeats.com/b/your-unique-id",
		})
	})

	// Heartbeat endpoint
	// Apply rate limiting: 3 req/min
	r.GET("/b/:uuid", middleware.RateLimitMiddleware(), handler.HandleHeartbeat)
	r.HEAD("/b/:uuid", middleware.RateLimitMiddleware(), handler.HandleHeartbeat)
	r.POST("/b/:uuid", middleware.RateLimitMiddleware(), handler.HandleHeartbeat)

	// Health check for the API itself
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "API is running"})
	})

	return r
}
