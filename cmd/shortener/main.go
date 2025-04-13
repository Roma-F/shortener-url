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

	logger.Initialize("info")
	defer logger.Sugar.Sync()

	r := router.NewRouterHandler(cfg)

	loggerRouter := middleware.WithLogging(r, logger.Sugar)

	s := server.NewServer(loggerRouter, cfg)

	logger.Sugar.Infof("Server will run on %s", cfg.RunAddr)
	logger.Sugar.Infof("Base URL is %s", cfg.ShortURLAddr)

	err = s.ListenAndServe()

	if err != nil {
		panic(err)
	}

	defer s.Close()
}
