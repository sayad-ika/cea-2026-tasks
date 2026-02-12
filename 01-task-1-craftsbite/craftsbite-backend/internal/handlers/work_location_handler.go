package handlers

import (
	"craftsbite-backend/internal/models"
	"craftsbite-backend/internal/services"
	"craftsbite-backend/internal/utils"

	"github.com/gin-gonic/gin"
)

// WorkLocationHandler handles work location endpoints
type WorkLocationHandler struct {
	workLocationService services.WorkLocationService
}

// NewWorkLocationHandler creates a new work location handler
func NewWorkLocationHandler(workLocationService services.WorkLocationService) *WorkLocationHandler {
	return &WorkLocationHandler{
		workLocationService: workLocationService,
	}
}

// SetMyLocationRequest represents request body for setting own location
type SetMyLocationRequest struct {
	Date     string `json:"date" binding:"required"`
	Location string `json:"location" binding:"required"`
}

// OverrideUserLocationRequest represents request body for overriding another user's location
type OverrideUserLocationRequest struct {
	UserID   string `json:"user_id" binding:"required"`
	Date     string `json:"date" binding:"required"`
	Location string `json:"location" binding:"required"`
	Reason   string `json:"reason" binding:"required"`
}

// CreateGlobalWFHPolicyRequest represents request body for creating global WFH policy
type CreateGlobalWFHPolicyRequest struct {
	StartDate string `json:"start_date" binding:"required"`
	EndDate   string `json:"end_date" binding:"required"`
	Reason    string `json:"reason"`
}

// GetMyLocation returns effective work location for authenticated user on a date
// GET /api/v1/work-locations/me?date=YYYY-MM-DD
func (h *WorkLocationHandler) GetMyLocation(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, 401, "UNAUTHORIZED", "User not authenticated")
		return
	}

	date := c.Query("date")
	if date == "" {
		utils.ErrorResponse(c, 400, "VALIDATION_ERROR", "date query parameter is required")
		return
	}

	response, err := h.workLocationService.GetMyLocation(userID.(string), date)
	if err != nil {
		utils.ErrorResponse(c, 400, "GET_WORK_LOCATION_ERROR", err.Error())
		return
	}

	utils.SuccessResponse(c, 200, response, "Work location retrieved successfully")
}

// SetMyLocation sets explicit work location for authenticated user
// PUT /api/v1/work-locations/me
func (h *WorkLocationHandler) SetMyLocation(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, 401, "UNAUTHORIZED", "User not authenticated")
		return
	}

	var req SetMyLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, 400, "VALIDATION_ERROR", "Invalid request body: "+err.Error())
		return
	}

	input := services.SetMyLocationInput{
		Date:     req.Date,
		Location: models.WorkLocation(req.Location),
	}

	if err := h.workLocationService.SetMyLocation(userID.(string), input); err != nil {
		utils.ErrorResponse(c, 400, "SET_WORK_LOCATION_ERROR", err.Error())
		return
	}

	utils.SuccessResponse(c, 200, nil, "Work location updated successfully")
}

// OverrideUserLocation updates another user's location (Team Lead/Admin)
// PUT /api/v1/work-locations/override
func (h *WorkLocationHandler) OverrideUserLocation(c *gin.Context) {
	requesterID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, 401, "UNAUTHORIZED", "User not authenticated")
		return
	}

	var req OverrideUserLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, 400, "VALIDATION_ERROR", "Invalid request body: "+err.Error())
		return
	}

	input := services.OverrideUserLocationInput{
		UserID:   req.UserID,
		Date:     req.Date,
		Location: models.WorkLocation(req.Location),
		Reason:   req.Reason,
	}

	if err := h.workLocationService.OverrideUserLocation(requesterID.(string), input); err != nil {
		utils.ErrorResponse(c, 400, "OVERRIDE_WORK_LOCATION_ERROR", err.Error())
		return
	}

	utils.SuccessResponse(c, 200, nil, "Work location overridden successfully")
}

// GetTeamLocations returns effective locations for teams led by authenticated team lead
// GET /api/v1/work-locations/team?date=YYYY-MM-DD
func (h *WorkLocationHandler) GetTeamLocations(c *gin.Context) {
	teamLeadID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, 401, "UNAUTHORIZED", "User not authenticated")
		return
	}

	date := c.Query("date")
	if date == "" {
		utils.ErrorResponse(c, 400, "VALIDATION_ERROR", "date query parameter is required")
		return
	}

	response, err := h.workLocationService.GetTeamLocations(teamLeadID.(string), date)
	if err != nil {
		utils.ErrorResponse(c, 400, "GET_TEAM_WORK_LOCATION_ERROR", err.Error())
		return
	}

	utils.SuccessResponse(c, 200, response, "Team work locations retrieved successfully")
}

// CreateGlobalWFHPolicy creates a global WFH policy (Admin/Logistics)
// POST /api/v1/work-locations/global-wfh
func (h *WorkLocationHandler) CreateGlobalWFHPolicy(c *gin.Context) {
	requesterID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, 401, "UNAUTHORIZED", "User not authenticated")
		return
	}

	var req CreateGlobalWFHPolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, 400, "VALIDATION_ERROR", "Invalid request body: "+err.Error())
		return
	}

	input := services.CreateGlobalWFHPolicyInput{
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
		Reason:    req.Reason,
	}

	policy, err := h.workLocationService.CreateGlobalWFHPolicy(requesterID.(string), input)
	if err != nil {
		utils.ErrorResponse(c, 400, "CREATE_GLOBAL_WFH_POLICY_ERROR", err.Error())
		return
	}

	utils.SuccessResponse(c, 201, policy, "Global WFH policy created successfully")
}

// GetGlobalWFHPolicies lists global WFH policies
// GET /api/v1/work-locations/global-wfh?start_date=YYYY-MM-DD&end_date=YYYY-MM-DD
func (h *WorkLocationHandler) GetGlobalWFHPolicies(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	policies, err := h.workLocationService.GetGlobalWFHPolicies(startDate, endDate)
	if err != nil {
		utils.ErrorResponse(c, 400, "GET_GLOBAL_WFH_POLICIES_ERROR", err.Error())
		return
	}

	utils.SuccessResponse(c, 200, policies, "Global WFH policies retrieved successfully")
}

// DeleteGlobalWFHPolicy deactivates a global WFH policy (Admin/Logistics)
// DELETE /api/v1/work-locations/global-wfh/:id
func (h *WorkLocationHandler) DeleteGlobalWFHPolicy(c *gin.Context) {
	requesterID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, 401, "UNAUTHORIZED", "User not authenticated")
		return
	}

	policyID := c.Param("id")
	if policyID == "" {
		utils.ErrorResponse(c, 400, "VALIDATION_ERROR", "policy ID is required")
		return
	}

	if err := h.workLocationService.DeleteGlobalWFHPolicy(requesterID.(string), policyID); err != nil {
		utils.ErrorResponse(c, 400, "DELETE_GLOBAL_WFH_POLICY_ERROR", err.Error())
		return
	}

	utils.SuccessResponse(c, 200, nil, "Global WFH policy deactivated successfully")
}
