package middleware

import (
	"craftsbite-backend/internal/models"
	"craftsbite-backend/internal/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates JWT tokens
func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.ErrorResponse(c, 401, "UNAUTHORIZED", "Authorization header required")
			c.Abort()
			return
		}

		// Check Bearer prefix
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.ErrorResponse(c, 401, "UNAUTHORIZED", "Invalid authorization header format")
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Validate token
		claims, err := utils.ValidateToken(tokenString, jwtSecret)
		if err != nil {
			utils.ErrorResponse(c, 401, "UNAUTHORIZED", "Invalid or expired token")
			c.Abort()
			return
		}

		// Set user claims in context
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)

		c.Next()
	}
}

// RequireRoles checks if the user has one of the required roles
func RequireRoles(roles ...models.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			utils.ErrorResponse(c, 403, "FORBIDDEN", "User role not found in context")
			c.Abort()
			return
		}

		userRoleStr := userRole.(string)

		// Check if user role matches any of the required roles
		hasPermission := false
		for _, role := range roles {
			if userRoleStr == role.String() {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			utils.ErrorResponse(c, 403, "FORBIDDEN", "Insufficient permissions")
			c.Abort()
			return
		}

		c.Next()
	}
}
