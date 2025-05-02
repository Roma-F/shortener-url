package router

import (
	"net/http"

	"github.com/Roma-F/shortener-url/internal/app/config"
	"github.com/Roma-F/shortener-url/internal/app/logger"
	"github.com/Roma-F/shortener-url/internal/app/service"
	"github.com/Roma-F/shortener-url/internal/app/storage"
	"github.com/Roma-F/shortener-url/internal/app/transport/handler"
	"github.com/go-chi/chi/v5"
)

func NewRouterHandler(cfg *config.ServerOption) http.Handler {
	r := chi.NewRouter()

	repo := storage.NewMemoryStorage(cfg.FSPath)
	db, err := storage.NewPostgresStorage(cfg.DatabaseDSN)
	if err != nil {
		logger.Sugar.Infow("Init db error")
	}

	URLService := service.NewURLService(repo, cfg)
	URLHandler := handler.NewURLHandler(URLService)
	LiveHandler := handler.NewPingHandler(db)

	r.Group(func(r chi.Router) {
		r.Post("/", URLHandler.ShortenURLTextPlain)
		r.Get("/{id}", URLHandler.GetMainURL)
		r.Get("/ping", LiveHandler.Ping)
	})

	r.Route("/api", func(r chi.Router) {
		r.Post("/shorten", URLHandler.ShortenURLJSON)
	})

	return r
}
