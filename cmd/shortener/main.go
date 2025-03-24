package main

import (
	"github.com/Roma-F/shortener-url/internal/app/router"
	"github.com/Roma-F/shortener-url/internal/app/server"
)

func main() {
	handler := router.NewRouterHandler()
	s := server.NewServer(handler, ":8080")

	err := s.ListenAndServe()

	if err != nil {
		panic(err)
	}

	defer s.Close()
}
