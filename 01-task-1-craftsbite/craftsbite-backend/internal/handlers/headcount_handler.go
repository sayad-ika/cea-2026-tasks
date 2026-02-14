package handlers

import (
	"craftsbite-backend/internal/services"
	"craftsbite-backend/internal/sse"
	"craftsbite-backend/internal/utils"
	"encoding/json"
	"time"

	"github.com/gin-gonic/gin"
)

// HeadcountHandler handles headcount reporting endpoints
type HeadcountHandler struct {
	headcountService services.HeadcountService
	sseHub           *sse.Hub
}

// NewHeadcountHandler creates a new headcount handler
func NewHeadcountHandler(headcountService services.HeadcountService, sseHub *sse.Hub) *HeadcountHandler {
	return &HeadcountHandler{
		headcountService: headcountService,
		sseHub:           sseHub,
	}
}

// GetTodayHeadcount returns today's and tomorrow's headcount summary
// GET /api/headcount/today
func (h *HeadcountHandler) GetTodayHeadcount(c *gin.Context) {
	summary, err := h.headcountService.GetTodayHeadcount()
	if err != nil {
		utils.ErrorResponse(c, 500, "INTERNAL_ERROR", err.Error())
		return
	}

	utils.SuccessResponse(c, 200, summary, "Today's and tomorrow's headcount retrieved successfully")
}

// GetHeadcountByDate returns headcount summary for a specific date
// GET /api/headcount/:date
func (h *HeadcountHandler) GetHeadcountByDate(c *gin.Context) {
	// Get date from URL parameter
	date := c.Param("date")
	if date == "" {
		utils.ErrorResponse(c, 400, "VALIDATION_ERROR", "Date parameter is required")
		return
	}

	summary, err := h.headcountService.GetHeadcountByDate(date)
	if err != nil {
		utils.ErrorResponse(c, 400, "VALIDATION_ERROR", err.Error())
		return
	}

	utils.SuccessResponse(c, 200, summary, "Headcount retrieved successfully")
}

// GetDetailedHeadcount returns detailed headcount for a specific date and meal
// GET /api/headcount/:date/:meal_type
func (h *HeadcountHandler) GetDetailedHeadcount(c *gin.Context) {
	// Get parameters from URL
	date := c.Param("date")
	mealType := c.Param("meal_type")

	if date == "" || mealType == "" {
		utils.ErrorResponse(c, 400, "VALIDATION_ERROR", "Date and meal_type parameters are required")
		return
	}

	details, err := h.headcountService.GetDetailedHeadcount(date, mealType)
	if err != nil {
		utils.ErrorResponse(c, 400, "VALIDATION_ERROR", err.Error())
		return
	}

	utils.SuccessResponse(c, 200, details, "Detailed headcount retrieved successfully")
}

// GetDailyAnnouncement returns a copy/paste-friendly daily announcement draft
// GET /api/headcount/announcement/:date
// GET /api/headcount/announcement?date=YYYY-MM-DD
func (h *HeadcountHandler) GetDailyAnnouncement(c *gin.Context) {
	date := c.Param("date")
	if date == "" {
		date = c.Query("date")
	}
	if date == "" {
		utils.ErrorResponse(c, 400, "VALIDATION_ERROR", "Date is required as path param '/announcement/:date' or query param '?date=YYYY-MM-DD'")
		return
	}

	announcement, err := h.headcountService.GenerateDailyAnnouncement(date)
	if err != nil {
		utils.ErrorResponse(c, 400, "VALIDATION_ERROR", err.Error())
		return
	}

	utils.SuccessResponse(c, 200, announcement, "Daily announcement draft generated successfully")
}

// GetEnhancedHeadcountReport returns improved headcount reporting for a date
// GET /api/headcount/report/:date
// GET /api/headcount/report?date=YYYY-MM-DD
func (h *HeadcountHandler) GetEnhancedHeadcountReport(c *gin.Context) {
	today := time.Now().Format("2006-01-02")
	tomorrow := time.Now().AddDate(0, 0, 1).Format("2006-01-02")

	reports, err := h.headcountService.GetEnhancedHeadcountReport(today, tomorrow)
	if err != nil {
		utils.ErrorResponse(c, 400, "VALIDATION_ERROR", err.Error())
		return
	}

	utils.SuccessResponse(c, 200, reports, "Today's and tomorrow's headcount retrieved successfully")
}

// BroadcastUpdatedHeadcount recalculates and pushes fresh data to all SSE clients.
// Call this from any handler that mutates participation data.
func (h *HeadcountHandler) BroadcastUpdatedHeadcount() {
	today := time.Now().Format("2006-01-02")
	tomorrow := time.Now().AddDate(0, 0, 1).Format("2006-01-02")

	reports, err := h.headcountService.GetEnhancedHeadcountReport(today, tomorrow)
	if err != nil {
		return
	}

	payload, err := json.Marshal(map[string]any{
		"success": true,
		"data":    reports,
		"message": "Today's and tomorrow's headcount retrieved successfully",
	})
	if err != nil {
		return
	}

	h.sseHub.BroadcastHeadcount(payload)
}
