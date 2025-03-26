package router

import (
	"net/http"

	"github.com/Roma-F/shortener-url/internal/app/config"
	"github.com/Roma-F/shortener-url/internal/app/handler"
	"github.com/Roma-F/shortener-url/internal/app/service"
	"github.com/Roma-F/shortener-url/internal/app/storage"
	"github.com/go-chi/chi/v5"
)

func NewRouterHandler(cfg *config.ServerOption) http.Handler {
	r := chi.NewRouter()

	repo := storage.NewMemoryStorage()
	URLService := service.NewURLService(repo, cfg)
	URLHandler := handler.NewURLHandler(URLService)

	r.Post("/", URLHandler.ShortenURL)
	r.Get("/{id}", URLHandler.GetMainURL)

	return r
}
