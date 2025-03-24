package server

import "net/http"

func NewServer(handler http.Handler, port string) *http.Server {
	s := http.Server{
		Addr:    ":" + port,
		Handler: handler,
	}

	return &s
}
