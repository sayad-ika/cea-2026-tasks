package handlers

import (
	"io"

	"craftsbite-backend/internal/models"
	"craftsbite-backend/internal/sse"
	"craftsbite-backend/internal/utils"

	"github.com/gin-gonic/gin"
)

// HeadcountSSEHandler streams live headcount updates to connected clients.
// GET /api/v1/headcount/report/live

// Uses the same AuthMiddleware as all other routes — cookie is sent
// automatically by the browser with withCredentials: true on the EventSource.
func HeadcountSSEHandler(hub *sse.Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Role guard — only admin and logistics can see headcount
		role := c.GetString("role")
		if role != models.RoleAdmin.String() && role != models.RoleLogistics.String() {
			utils.ErrorResponse(c, 403, "FORBIDDEN", "Insufficient permissions")
			return
		}

		// SSE headers — must be set before any write
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("X-Accel-Buffering", "no") // critical if behind nginx

		// Register this browser tab with the hub
		client := &sse.Client{
			Channel: make(chan []byte, 10),
		}
		hub.Register <- client

		// Unregister when the browser tab closes or navigates away
		defer func() {
			hub.Unregister <- client
		}()

		// Stream events until client disconnects
		c.Stream(func(w io.Writer) bool {
			select {
			case payload, ok := <-client.Channel:
				if !ok {
					return false
				}
				c.SSEvent("headcount-update", string(payload))
				return true

			case <-c.Request.Context().Done():
				return false
			}
		})
	}
}
