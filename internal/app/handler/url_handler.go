package handler

import (
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

type URLShortener interface {
	FetchOriginalURL(id string) (string, error)
	GenerateShortURL(originalURL string) (string, error)
}

type URLHandler struct {
	service URLShortener
}

func NewURLHandler(svc URLShortener) *URLHandler {
	return &URLHandler{service: svc}
}

func (h *URLHandler) ShortenURL(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "text/plain") {
		http.Error(w, "Content-Type most be text/plain", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	url := string(body)
	shortURL, err := h.service.GenerateShortURL(url)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Length", strconv.Itoa(len(shortURL)))
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}

func (h *URLHandler) GetMainURL(w http.ResponseWriter, r *http.Request) {
	urlID := chi.URLParam(r, "id")

	mainURL, err := h.service.FetchOriginalURL(urlID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	http.Redirect(w, r, mainURL, http.StatusTemporaryRedirect)
}
