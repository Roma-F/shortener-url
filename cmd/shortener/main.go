package main

import (
	"log"

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

	r := router.NewRouterHandler(cfg)

	gzipRouter := middleware.WithGzip(r)
	loggerRouter := middleware.WithLogging(gzipRouter, logger.Sugar)

	s := server.NewServer(loggerRouter, cfg)

	defer func() {
		logger.Sugar.Info("Server stopping...")

		if err := s.Close(); err != nil {
			logger.Sugar.Errorw("Failed to close server", "error", err)
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
