package handler

import (
	"net/http"

	"github.com/Roma-F/shortener-url/internal/app/storage"
)

type PingHandler struct {
	db *storage.PostgresStorage
}

func NewPingHandler(db *storage.PostgresStorage) *PingHandler {
	return &PingHandler{
		db: db,
	}
}

func (h *PingHandler) Ping(w http.ResponseWriter, r *http.Request) {
	if h.db == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := h.db.Ping(r.Context()); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
