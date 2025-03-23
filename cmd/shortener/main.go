package main

import (
	"fmt"
	"net/http"

	"github.com/Roma-F/shortener-url/internal/app/handler"
	"github.com/Roma-F/shortener-url/internal/app/service"
	"github.com/Roma-F/shortener-url/internal/app/storage"
)

func main() {
	repo := storage.NewMemoryStorage()
	URLService := service.NewURLService(repo)
	URLHandler := handler.NewURLHandler(URLService)

	mux := http.NewServeMux()
	mux.HandleFunc("/", URLHandler.ShortenURL)
	mux.HandleFunc("/{id}", URLHandler.GetMainURL)

	fmt.Println("Server is running on :8080")
	err := http.ListenAndServe(":8080", mux)
	fmt.Println()
	if err != nil {
		panic(err)
	}
}
