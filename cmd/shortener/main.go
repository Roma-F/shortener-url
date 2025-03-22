package main

import (
	"fmt"
	"net/http"

	"github.com/Roma-F/shortener-url/internal/app/handler"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler.ShortenURL)
	mux.HandleFunc("/{id}", handler.GetMainURL)

	fmt.Println("Server is running on :8080")
	err := http.ListenAndServe(`:8080`, mux)
	fmt.Println()
	if err != nil {
		panic(err)
	}
}
