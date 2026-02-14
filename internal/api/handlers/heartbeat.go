package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/uptimebeats/heartbeat-api/internal/models"
	"github.com/uptimebeats/heartbeat-api/internal/repository"
)

type HeartbeatHandler struct {
	Repo *repository.HeartbeatRepository
}

func NewHeartbeatHandler(repo *repository.HeartbeatRepository) *HeartbeatHandler {
	return &HeartbeatHandler{Repo: repo}
}

// HandleHeartbeat processes the incoming heartbeat request
// GET /b/:uuid
func (h *HeartbeatHandler) HandleHeartbeat(c *gin.Context) {
	uniqueID := c.Param("uuid")
	if uniqueID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing unique ID"})
		return
	}

	ctx := c.Request.Context()

	// 1. Verify existence of the site
	site, err := h.Repo.GetSiteByUniqueID(ctx, uniqueID)
	if err != nil {
		log.Printf("Error checking site %s: %v", uniqueID, err)
		c.Status(http.StatusInternalServerError)
		return
	}
	if site == nil {
		c.Status(http.StatusNotFound)
		return
	}

	// 2. Record heartbeat
	// We extract some metadata from the request
	userAgent := c.Request.UserAgent()
	clientIP := c.ClientIP()
	method := c.Request.Method
	now := time.Now()
	// Since this is a heartbeat check *from* the user's service calling *us*, the status code is technically 200 from *our* perspective.
	// But the schema has `status_code` column. Does this mean the status code WE return, or the status code of the service?
	// The prompt says "which will get pinged from url like below".
	// Usually this means the user's service calls this URL to say "I'm alive".
	// The `status_code` in `sites_status_heartbeat` might be for *our* response or just placeholder.
	// Let's assume 200 for now as it's a successful ping received.
	statusCode := int64(http.StatusOK)

	status := &models.HeartbeatStatus{
		HeartbeatSiteID: site.ID,
		ReceivedAt:      &now,
		SourceIP:        &clientIP,
		HTTPMethod:      &method,
		UserAgent:       &userAgent,
		StatusCode:      &statusCode,
		Error:           nil, // No error, it's a success
	}

	if err := h.Repo.RecordHeartbeat(ctx, status); err != nil {
		log.Printf("Error recording heartbeat for %s: %v", uniqueID, err)
		c.Status(http.StatusInternalServerError)
		return
	}

	// 3. Return success
	// Return a small JSON or just 200 OK.
	// Many heartbeat services return a small image or JSON. JSON is safer.
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
