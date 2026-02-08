package main

import (
	"fmt"
	"log"

	"craftsbite-backend/internal/config"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Print configuration (for verification during development)
	if cfg.IsDevelopment() {
		fmt.Println("=================================")
		fmt.Println("CraftsBite Backend Configuration")
		fmt.Println("=================================")
		fmt.Printf("Environment: %s\n", cfg.Server.Env)
		fmt.Printf("Server Address: %s\n", cfg.Server.GetServerAddress())
		fmt.Printf("Database: %s@%s:%s/%s\n",
			cfg.Database.User,
			cfg.Database.Host,
			cfg.Database.Port,
			cfg.Database.Name,
		)
		fmt.Printf("JWT Expiration: %s\n", cfg.JWT.Expiration)
		fmt.Printf("CORS Allowed Origins: %v\n", cfg.CORS.AllowedOrigins)
		fmt.Printf("Log Level: %s\n", cfg.Logging.Level)
		fmt.Printf("Meal Cutoff Time: %s (%s)\n", cfg.Meal.CutoffTime, cfg.Meal.CutoffTimezone)
		fmt.Printf("History Retention: %d months\n", cfg.Cleanup.RetentionMonths)
		fmt.Printf("Rate Limiting: %t (%d req/min)\n", cfg.RateLimit.Enabled, cfg.RateLimit.RequestsPerMinute)
		fmt.Println("=================================\n")
	}

	// Set Gin mode based on environment
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize router
	router := gin.Default()

	// Basic ping endpoint
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		v1.GET("/health", healthCheck(cfg))
	}

	// Start server
	addr := cfg.Server.GetServerAddress()
	log.Printf("Starting CraftsBite API server on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func healthCheck(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":      "healthy",
			"service":     "craftsbite-api",
			"environment": cfg.Server.Env,
		})
	}
}
