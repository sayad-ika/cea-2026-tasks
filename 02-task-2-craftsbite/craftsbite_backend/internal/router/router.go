package router

import (
	"net/http"

	"craftsbite-backend/internal/config"

	"github.com/gin-gonic/gin"
)

func New(cfg *config.Config) *gin.Engine {
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.POST("/discord/interact", func(c *gin.Context) {
		c.Status(http.StatusNotImplemented)
	})

	return r
}
