package main

import (
	"log"

	"github.com/Roma-F/shortener-url/internal/app/config"
	"github.com/Roma-F/shortener-url/internal/app/router"
	"github.com/Roma-F/shortener-url/internal/app/server"
)

func main() {
	cfg, err := config.NewServerOption()
	if err != nil {
		log.Fatalf("Failed to load server configuration: %v", err)
	}
	handler := router.NewRouterHandler(cfg)

	s := server.NewServer(handler, cfg)

	log.Printf("Server will run on %s", cfg.RunAddr)
	log.Printf("Base URL is %s", cfg.ShortURLAddr)

	err = s.ListenAndServe()

	if err != nil {
		panic(err)
	}

	defer s.Close()
}
