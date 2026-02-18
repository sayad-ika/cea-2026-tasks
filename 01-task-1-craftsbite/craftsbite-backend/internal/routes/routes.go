package routes

import (
    "craftsbite-backend/internal/config"
    "craftsbite-backend/internal/handlers"
    "craftsbite-backend/internal/middleware"
    "craftsbite-backend/internal/models"

    "github.com/gin-gonic/gin"
)

type Handlers struct {
    Auth        *handlers.AuthHandler
    User        *handlers.UserHandler
    Meal        *handlers.MealHandler
    Schedule    *handlers.ScheduleHandler
    Headcount   *handlers.HeadcountHandler
    Preference  *handlers.PreferenceHandler
    BulkOptOut  *handlers.BulkOptOutHandler
    History     *handlers.HistoryHandler
}

func RegisterRoutes(router *gin.Engine, h *Handlers, cfg *config.Config) {
    // Health check endpoint (public)
    router.GET("/health", healthCheck(cfg))

    v1 := router.Group("/api/v1")
    {
        registerAuthRoutes(v1, h, cfg)
        registerUserRoutes(v1, h, cfg)
        registerMealRoutes(v1, h, cfg)
        registerScheduleRoutes(v1, h, cfg)
        registerHeadcountRoutes(v1, h, cfg)
        registerAdminRoutes(v1, h, cfg)
    }
}

func registerAuthRoutes(v1 *gin.RouterGroup, h *Handlers, cfg *config.Config) {
    // Public auth routes
    auth := v1.Group("/auth")
    {
        auth.POST("/login", h.Auth.Login)
        auth.POST("/register", h.Auth.Register)
    }

    // Protected auth routes
    authProtected := v1.Group("/auth")
    authProtected.Use(middleware.AuthMiddleware(cfg.JWT.Secret))
    {
        authProtected.GET("/me", h.Auth.GetCurrentUser)
        authProtected.POST("/logout", h.Auth.Logout)
    }
}

func registerUserRoutes(v1 *gin.RouterGroup, h *Handlers, cfg *config.Config) {
    users := v1.Group("/users")
    users.Use(middleware.AuthMiddleware(cfg.JWT.Secret))
    {
        // Admin only routes
        users.GET("", middleware.RequireRoles(models.RoleAdmin), h.User.ListUsers)
        users.POST("", middleware.RequireRoles(models.RoleAdmin), h.User.CreateUser)
        users.DELETE("/:id", middleware.RequireRoles(models.RoleAdmin), h.User.DeactivateUser)

        // Admin or Self routes
        users.GET("/:id", h.User.GetUser)
        users.PUT("/:id", h.User.UpdateUser)

        // Preference routes
        users.GET("/me/preferences", h.Preference.GetPreferences)
        users.PUT("/me/preferences", h.Preference.UpdatePreferences)

        // Team Lead routes
        users.GET("/me/team-members", middleware.RequireRoles(models.RoleTeamLead), h.User.GetMyTeamMembers)

        users.GET("/me/team", middleware.RequireRoles(models.RoleEmployee, models.RoleTeamLead), h.User.GetMyTeam)
    }
}

func registerMealRoutes(v1 *gin.RouterGroup, h *Handlers, cfg *config.Config) {
    meals := v1.Group("/meals")
    meals.Use(middleware.AuthMiddleware(cfg.JWT.Secret))
    {
        // User routes
        meals.GET("/today", h.Meal.GetTodayMeals)
        meals.GET("/participation/:date", h.Meal.GetParticipationByDate)
        meals.POST("/participation", h.Meal.SetParticipation)

        // Override routes
        meals.POST("/participation/override", middleware.RequireRoles(models.RoleAdmin, models.RoleTeamLead), h.Meal.OverrideParticipation)

        // Bulk opt-out routes
        meals.GET("/bulk-optouts", h.BulkOptOut.GetBulkOptOuts)
        meals.POST("/bulk-optouts", h.BulkOptOut.CreateBulkOptOut)
        meals.DELETE("/bulk-optouts/:id", h.BulkOptOut.DeleteBulkOptOut)

        // History routes
        meals.GET("/history", h.History.GetHistory)
        meals.GET("/participation-audit", h.History.GetAuditTrail)

        meals.GET("/team-participation", middleware.RequireRoles(models.RoleTeamLead), h.Meal.GetTeamParticipation)
        meals.GET("/all-teams-participation", middleware.RequireRoles(models.RoleAdmin, models.RoleLogistics), h.Meal.GetAllTeamsParticipation)
    }
}

func registerScheduleRoutes(v1 *gin.RouterGroup, h *Handlers, cfg *config.Config) {
    schedules := v1.Group("/schedules")
    schedules.Use(middleware.AuthMiddleware(cfg.JWT.Secret))
    {
        // Read routes - all authenticated users
        schedules.GET("/:date", h.Schedule.GetSchedule)
        schedules.GET("/range", h.Schedule.GetScheduleRange)

        // Write routes - admin only
        schedules.POST("", middleware.RequireRoles(models.RoleAdmin), h.Schedule.CreateSchedule)
        schedules.PUT("/:date", middleware.RequireRoles(models.RoleAdmin), h.Schedule.UpdateSchedule)
        schedules.DELETE("/:date", middleware.RequireRoles(models.RoleAdmin), h.Schedule.DeleteSchedule)
    }
}

func registerHeadcountRoutes(v1 *gin.RouterGroup, h *Handlers, cfg *config.Config) {
    headcount := v1.Group("/headcount")
    headcount.Use(middleware.AuthMiddleware(cfg.JWT.Secret))
    headcount.Use(middleware.RequireRoles(models.RoleAdmin, models.RoleLogistics))
    {
        headcount.GET("/today", h.Headcount.GetTodayHeadcount)
        headcount.GET("/:date", h.Headcount.GetHeadcountByDate)
        headcount.GET("/:date/:meal_type", h.Headcount.GetDetailedHeadcount)
    }
}

func registerAdminRoutes(v1 *gin.RouterGroup, h *Handlers, cfg *config.Config) {
    admin := v1.Group("/admin")
    admin.Use(middleware.AuthMiddleware(cfg.JWT.Secret))
    admin.Use(middleware.RequireRoles(models.RoleAdmin, models.RoleTeamLead))
    {
        admin.GET("/meals/history/:user_id", h.History.GetUserHistoryAdmin)
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