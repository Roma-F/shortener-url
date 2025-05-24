package router

import (
	"net/http"

	"github.com/Roma-F/shortener-url/internal/app/config"
	"github.com/Roma-F/shortener-url/internal/app/service"
	"github.com/Roma-F/shortener-url/internal/app/transport/handler"
	"github.com/Roma-F/shortener-url/internal/app/transport/middleware"
	"github.com/go-chi/chi/v5"
)

func NewRouter(urlService *service.URLService, healthService *service.HealthService, cfg *config.ServerOption) http.Handler {
	r := chi.NewRouter()

	URLHandler := handler.NewURLHandler(urlService)
	LiveHandler := handler.NewPingHandler(healthService)

	r.Use(middleware.WithAuthentication(cfg.AuthSecretKey))

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

		r.Route("/user", func(r chi.Router) {
			r.Get("/urls", URLHandler.GetUserURLs)
		})
	})

	return r
}
