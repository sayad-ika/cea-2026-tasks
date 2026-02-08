package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"craftsbite-backend/internal/config"
	"craftsbite-backend/internal/database"
	"craftsbite-backend/internal/handlers"
	"craftsbite-backend/internal/middleware"
	"craftsbite-backend/internal/models"
	"craftsbite-backend/internal/repository"
	"craftsbite-backend/internal/services"
	"craftsbite-backend/pkg/logger"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	logFormat := "console"
	if cfg.IsProduction() {
		logFormat = "json"
	}
	if err := logger.Init(cfg.Logging.Level, logFormat); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

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
		fmt.Println("=================================\n")
	}

	// Initialize database
	db, err := database.Connect(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close(db)

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	mealRepo := repository.NewMealRepository(db)
	scheduleRepo := repository.NewScheduleRepository(db)
	bulkOptOutRepo := repository.NewBulkOptOutRepository(db)
	historyRepo := repository.NewHistoryRepository(db)
	teamRepo := repository.NewTeamRepository(db)

	// Initialize services
	authService := services.NewAuthService(userRepo, cfg)
	userService := services.NewUserService(userRepo)
	participationResolver := services.NewParticipationResolver(mealRepo, scheduleRepo, bulkOptOutRepo, userRepo, cfg)
	mealService := services.NewMealService(mealRepo, scheduleRepo, historyRepo, userRepo, teamRepo, participationResolver, cfg)
	scheduleService := services.NewScheduleService(scheduleRepo)
	headcountService := services.NewHeadcountService(userRepo, scheduleRepo, participationResolver)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService)
	mealHandler := handlers.NewMealHandler(mealService)
	scheduleHandler := handlers.NewScheduleHandler(scheduleService)
	headcountHandler := handlers.NewHeadcountHandler(headcountService)

	// Set Gin mode based on environment
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize router without default middleware
	router := gin.New()

	// Apply global middleware in order
	router.Use(middleware.RecoveryMiddleware())
	router.Use(middleware.RequestIDMiddleware())
	router.Use(middleware.LoggerMiddleware())
	router.Use(middleware.CORSMiddleware(cfg.CORS.AllowedOrigins))

	// Health check endpoint (public)
	router.GET("/health", healthCheck(cfg))

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Public auth routes
		auth := v1.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
		}

		// Protected auth routes
		authProtected := v1.Group("/auth")
		authProtected.Use(middleware.AuthMiddleware(cfg.JWT.Secret))
		{
			authProtected.GET("/me", authHandler.GetCurrentUser)
			authProtected.POST("/logout", authHandler.Logout)
		}

		// Protected user routes
		users := v1.Group("/users")
		users.Use(middleware.AuthMiddleware(cfg.JWT.Secret))
		{
			// Admin only routes
			users.GET("", middleware.RequireRoles(models.RoleAdmin), userHandler.ListUsers)
			users.POST("", middleware.RequireRoles(models.RoleAdmin), userHandler.CreateUser)
			users.DELETE("/:id", middleware.RequireRoles(models.RoleAdmin), userHandler.DeactivateUser)

			// Admin or Self routes
			users.GET("/:id", userHandler.GetUser)
			users.PUT("/:id", userHandler.UpdateUser)
		}

		// Protected meal routes
		meals := v1.Group("/meals")
		meals.Use(middleware.AuthMiddleware(cfg.JWT.Secret))
		{
			// User routes - any authenticated user
			meals.GET("/today", mealHandler.GetTodayMeals)
			meals.GET("/participation/:date", mealHandler.GetParticipationByDate)
			meals.POST("/participation", mealHandler.SetParticipation)

			// Override routes - Admin, Team Lead, or Logistics
			meals.POST("/participation/override", middleware.RequireRoles(models.RoleAdmin, models.RoleTeamLead, models.RoleLogistics), mealHandler.OverrideParticipation)
		}

		// Protected schedule routes (Admin only)
		schedules := v1.Group("/schedules")
		schedules.Use(middleware.AuthMiddleware(cfg.JWT.Secret))
		{
			// Read routes - all authenticated users
			schedules.GET("/:date", scheduleHandler.GetSchedule)
			schedules.GET("/range", scheduleHandler.GetScheduleRange)

			// Write routes - admin only
			schedules.POST("", middleware.RequireRoles(models.RoleAdmin), scheduleHandler.CreateSchedule)
			schedules.PUT("/:date", middleware.RequireRoles(models.RoleAdmin), scheduleHandler.UpdateSchedule)
			schedules.DELETE("/:date", middleware.RequireRoles(models.RoleAdmin), scheduleHandler.DeleteSchedule)
		}

		// Protected headcount routes (Admin only)
		headcount := v1.Group("/headcount")
		headcount.Use(middleware.AuthMiddleware(cfg.JWT.Secret))
		headcount.Use(middleware.RequireRoles(models.RoleAdmin))
		{
			headcount.GET("/today", headcountHandler.GetTodayHeadcount)
			headcount.GET("/:date", headcountHandler.GetHeadcountByDate)
			headcount.GET("/:date/:meal_type", headcountHandler.GetDetailedHeadcount)
		}
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:    cfg.Server.GetServerAddress(),
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		logger.Info(fmt.Sprintf("Starting CraftsBite API server on %s", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	// Graceful shutdown with 5 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Info("Server exited")
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
