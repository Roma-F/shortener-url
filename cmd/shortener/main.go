package main

import (
	"context"
	"log"
	"time"

	"github.com/Roma-F/shortener-url/internal/app/config"
	"github.com/Roma-F/shortener-url/internal/app/logger"
	"github.com/Roma-F/shortener-url/internal/app/router"
	"github.com/Roma-F/shortener-url/internal/app/server"
	"github.com/Roma-F/shortener-url/internal/app/transport/middleware"
)

func main() {
	cfg, err := config.NewServerOption()
	if err != nil {
		log.Fatalf("Failed to load server configuration: %v", err)
	}

	logger.Initialize(cfg.LoggingLevel)

	logger.Sugar.Infof("%s", cfg)

	r, pgStorage := router.NewRouterHandler(cfg)

	gzipRouter := middleware.WithGzip(r)
	loggerRouter := middleware.WithLogging(gzipRouter, logger.Sugar)

	s := server.NewServer(loggerRouter, cfg)

	defer func() {
		logger.Sugar.Info("Server stopping...")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := s.Shutdown(ctx); err != nil {
			logger.Sugar.Errorw("Failed to shutdown server gracefully", "error", err)
		}

		if pgStorage != nil {
			if err := pgStorage.Close(); err != nil {
				logger.Sugar.Errorw("Failed to close database connection", "error", err)
			} else {
				logger.Sugar.Info("Database connection closed")
			}
		}

		if err := logger.Sugar.Sync(); err != nil {
			log.Printf("Failed to sync logger buffers: %v", err)
		}

		logger.Sugar.Info("Server stopped")
	}()

	if err := s.ListenAndServe(); err != nil {
		logger.Sugar.Errorw("Server error", "error", err)
	}
}
