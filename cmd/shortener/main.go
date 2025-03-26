package main

import (
	"github.com/Roma-F/shortener-url/internal/app/config"
	"github.com/Roma-F/shortener-url/internal/app/router"
	"github.com/Roma-F/shortener-url/internal/app/server"
)

func main() {
	cfg := config.NewServerOption()
	handler := router.NewRouterHandler(cfg)

	s := server.NewServer(handler, cfg)

	err := s.ListenAndServe()

	if err != nil {
		panic(err)
	}

	defer s.Close()
}
