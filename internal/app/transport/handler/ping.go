package handler

import (
	"context"
	"net/http"
)

type HealthService interface {
	PingDB(ctx context.Context) error
}

type PingHandler struct {
	service HealthService
}

func NewPingHandler(service HealthService) *PingHandler {
	return &PingHandler{
		service: service,
	}
}

func (h *PingHandler) Ping(w http.ResponseWriter, r *http.Request) {
	if err := h.service.PingDB(r.Context()); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
