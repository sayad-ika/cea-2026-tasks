package handlers

import (
	"craftsbite-backend/internal/services"
	"craftsbite-backend/internal/utils"

	"github.com/gin-gonic/gin"
)

// LoginRequest represents login request body
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	authService services.AuthService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, 400, "VALIDATION_ERROR", err.Error())
		return
	}

	response, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		utils.ErrorResponse(c, 401, "INVALID_CREDENTIALS", err.Error())
		return
	}

	utils.SuccessResponse(c, 200, response, "Login successful")
}

// Logout handles user logout (placeholder)
func (h *AuthHandler) Logout(c *gin.Context) {
	// In a stateless JWT system, logout is typically handled client-side
	// by removing the token. This is a placeholder for future enhancements
	// like token blacklisting or refresh token revocation.
	utils.SuccessResponse(c, 200, nil, "Logout successful")
}

// GetCurrentUser returns the current authenticated user
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, 401, "UNAUTHORIZED", "User not authenticated")
		return
	}

	user, err := h.authService.GetCurrentUser(userID.(string))
	if err != nil {
		utils.ErrorResponse(c, 404, "USER_NOT_FOUND", err.Error())
		return
	}

	utils.SuccessResponse(c, 200, user, "User retrieved successfully")
}
