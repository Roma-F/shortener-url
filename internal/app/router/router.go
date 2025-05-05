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

func NewRouterHandler(cfg *config.ServerOption) (http.Handler, *storage.PostgresStorage) {
	r := chi.NewRouter()

	var repo service.Repository
	var dbStorage *storage.PostgresStorage
	var err error

	if cfg.DatabaseDSN != "" {
		dbStorage, err = storage.NewPostgresStorage(cfg.DatabaseDSN)
		if err == nil {
			logger.Sugar.Infow("Using PostgreSQL database for URL storage")
			repo = dbStorage
		} else {
			logger.Sugar.Warnw("Failed to connect to PostgreSQL database, falling back to file storage", "error", err)
		}
	}

	if repo == nil && cfg.FSPath != "" {
		logger.Sugar.Infow("Using file storage for URL storage", "path", cfg.FSPath)
		repo = storage.NewMemoryStorage(cfg.FSPath)
	}

	if repo == nil {
		logger.Sugar.Infow("Using in-memory storage for URL storage")
		repo = storage.NewMemoryStorage("")
	}

	URLService := service.NewURLService(repo, cfg)
	URLHandler := handler.NewURLHandler(URLService)
	LiveHandler := handler.NewPingHandler(dbStorage)

	r.Group(func(r chi.Router) {
		r.Post("/", URLHandler.ShortenURLTextPlain)
		r.Get("/{id}", URLHandler.GetMainURL)
		r.Get("/ping", LiveHandler.Ping)
	})

	r.Route("/api", func(r chi.Router) {
		r.Route("/shorten", func(r chi.Router) {
			r.Post("/", URLHandler.ShortenURLJSON)
			r.Post("/batch", URLHandler.ShortenURLBatch)
		})
	})

	return r, dbStorage
}
