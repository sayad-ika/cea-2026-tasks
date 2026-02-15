package handlers

import (
	"craftsbite-backend/internal/services"
	"craftsbite-backend/internal/utils"

	"github.com/gin-gonic/gin"
)

// BulkOptOutHandler handles bulk opt-out endpoints
type BulkOptOutHandler struct {
	bulkOptOutService services.BulkOptOutService
}

// NewBulkOptOutHandler creates a new bulk opt-out handler
func NewBulkOptOutHandler(bulkOptOutService services.BulkOptOutService) *BulkOptOutHandler {
	return &BulkOptOutHandler{
		bulkOptOutService: bulkOptOutService,
	}
}

// GetBulkOptOuts returns all bulk opt-outs for the current user
// GET /api/v1/meals/bulk-optouts
func (h *BulkOptOutHandler) GetBulkOptOuts(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, 401, "UNAUTHORIZED", "User not authenticated")
		return
	}

	optOuts, err := h.bulkOptOutService.GetBulkOptOuts(userID.(string))
	if err != nil {
		utils.ErrorResponse(c, 500, "INTERNAL_ERROR", err.Error())
		return
	}

	utils.SuccessResponse(c, 200, optOuts, "Bulk opt-outs retrieved successfully")
}

// CreateBulkOptOutRequest represents the request body for creating a bulk opt-out
type CreateBulkOptOutRequest struct {
	StartDate string `json:"start_date" binding:"required"`
	EndDate   string `json:"end_date" binding:"required"`
	MealType  string `json:"meal_type" binding:"required"`
	Reason    string `json:"reason"`
}

// CreateBatchBulkOptOutRequest represents the request body for creating a batch bulk opt-out
type CreateBatchBulkOptOutRequest struct {
	UserIDs   []string `json:"user_ids" binding:"required"`
	StartDate string   `json:"start_date" binding:"required"`
	EndDate   string   `json:"end_date" binding:"required"`
	MealType  string   `json:"meal_type" binding:"required"`
	Reason    string   `json:"reason"`
}

// CreateBulkOptOut creates a new bulk opt-out for the current user
// POST /api/v1/meals/bulk-optouts
func (h *BulkOptOutHandler) CreateBulkOptOut(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, 401, "UNAUTHORIZED", "User not authenticated")
		return
	}

	// Parse request body
	var req CreateBulkOptOutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, 400, "VALIDATION_ERROR", "Invalid request body: "+err.Error())
		return
	}

	// Create bulk opt-out
	input := services.CreateBulkOptOutInput{
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
		MealType:  req.MealType,
		Reason:    req.Reason,
	}

	// For single create, CreatedBy is the user themselves
	optOut, err := h.bulkOptOutService.CreateBulkOptOut(userID.(string), input, userID.(string))
	if err != nil {
		utils.ErrorResponse(c, 400, "CREATE_BULK_OPTOUT_ERROR", err.Error())
		return
	}

	utils.SuccessResponse(c, 201, optOut, "Bulk opt-out created successfully")
}

// DeleteBulkOptOut deletes a bulk opt-out
// DELETE /api/v1/meals/bulk-optouts/:id
func (h *BulkOptOutHandler) DeleteBulkOptOut(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, 401, "UNAUTHORIZED", "User not authenticated")
		return
	}

	// Get bulk opt-out ID from URL parameter
	id := c.Param("id")
	if id == "" {
		utils.ErrorResponse(c, 400, "VALIDATION_ERROR", "Bulk opt-out ID is required")
		return
	}

	// Delete bulk opt-out
	err := h.bulkOptOutService.DeleteBulkOptOut(userID.(string), id)
	if err != nil {
		utils.ErrorResponse(c, 400, "DELETE_BULK_OPTOUT_ERROR", err.Error())
		return
	}

	utils.SuccessResponse(c, 200, nil, "Bulk opt-out deleted successfully")
}

// CreateBatchBulkOptOut creates bulk opt-outs for multiple users
// POST /api/v1/meals/bulk-optouts/batch
func (h *BulkOptOutHandler) CreateBatchBulkOptOut(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, 401, "UNAUTHORIZED", "User not authenticated")
		return
	}

	// Parse request body
	var req CreateBatchBulkOptOutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, 400, "VALIDATION_ERROR", "Invalid request body: "+err.Error())
		return
	}

	// Create bulk opt-out input
	input := services.CreateBulkOptOutInput{
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
		MealType:  req.MealType,
		Reason:    req.Reason,
	}

	// Call service
	optOuts, err := h.bulkOptOutService.CreateBatchBulkOptOut(req.UserIDs, input, userID.(string))
	if err != nil {
		utils.ErrorResponse(c, 400, "CREATE_BATCH_BULK_OPTOUT_ERROR", err.Error())
		return
	}

	utils.SuccessResponse(c, 201, optOuts, "Batch bulk opt-outs created successfully")
}
