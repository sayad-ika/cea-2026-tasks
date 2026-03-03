package main

import (
	"fmt"
	"log"

	"craftsbite-backend/internal/config"
	"craftsbite-backend/internal/router"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	// .env is optional; ignored if not present (production uses injected env vars).
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to initialise logger: %v", err)
	}
	defer logger.Sync() //nolint:errcheck

	r := router.New(cfg)

	logger.Info("starting server", zap.Int("port", cfg.Server.Port))
	if err := r.Run(fmt.Sprintf(":%d", cfg.Server.Port)); err != nil {
		logger.Fatal("server failed", zap.Error(err))
	}
}
