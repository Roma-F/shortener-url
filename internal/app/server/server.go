package server

import (
	"net/http"

	"github.com/Roma-F/shortener-url/internal/app/config"
)

func NewServer(handler http.Handler, cfg *config.ServerOption) *http.Server {

	s := http.Server{
		Addr:    cfg.RunAddr,
		Handler: handler,
	}

	return &s
}
