package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/Roma-F/shortener-url/internal/app/logger"
	"github.com/Roma-F/shortener-url/internal/app/models"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type URLShortener interface {
	FetchOriginalURL(id string) (string, error)
	GenerateShortURL(originalURL string) (string, error)
	ShortenBatch(requests []models.ShortenBatchItem) ([]models.ShortenedURLItem, error)
}

type URLHandler struct {
	service URLShortener
}

func NewURLHandler(svc URLShortener) *URLHandler {
	return &URLHandler{service: svc}
}

func (h *URLHandler) ShortenURLTextPlain(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "" && !strings.HasPrefix(contentType, "text/plain") {
		http.Error(w, "Content-Type must be text/plain", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}

	defer func() {
		if err := r.Body.Close(); err != nil {
			logger.Sugar.Errorw("Failed to close request body", "error", err)
		}
	}()

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

func (h *URLHandler) ShortenURLJSON(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "" && !strings.HasPrefix(contentType, "application/json") {
		http.Error(w, "Content-Type must be application/json", http.StatusBadRequest)
		return
	}

	var req models.ShortenURLReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Sugar.Debug("cannot decode request JSON body", zap.Error(err))
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	defer func() {
		if err := r.Body.Close(); err != nil {
			logger.Sugar.Errorw("Failed to close request body", "error", err)
		}
	}()

	shortURL, err := h.service.GenerateShortURL(req.URL)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	resp := models.ShortenURLResp{
		Result: shortURL,
	}

	jsonData, err := json.MarshalIndent(resp, "", "   ")
	if err != nil {
		http.Error(w, "Error creating JSON response: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(jsonData)))
	w.WriteHeader(http.StatusCreated)
	w.Write(jsonData)

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

func (h *URLHandler) ShortenURLBatch(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "" && !strings.HasPrefix(contentType, "application/json") {
		http.Error(w, "Content-Type must be application/json", http.StatusBadRequest)
		return
	}

	var records []models.ShortenBatchItem
	if err := json.NewDecoder(r.Body).Decode(&records); err != nil {
		logger.Sugar.Debug("cannot decode request JSON body", zap.Error(err))
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	defer func() {
		if err := r.Body.Close(); err != nil {
			logger.Sugar.Errorw("Failed to close request body", "error", err)
		}
	}()

	if len(records) == 0 {
		http.Error(w, "Empty batch", http.StatusBadRequest)
		return
	}

	shortenBatch, err := h.service.ShortenBatch(records)
	if err != nil {
		logger.Sugar.Errorw("Failed to shorten URLs batch", "error", err)
		http.Error(w, "Failed to process batch", http.StatusInternalServerError)
		return
	}

	jsonData, err := json.MarshalIndent(shortenBatch, "", "   ")
	if err != nil {
		http.Error(w, "Error creating JSON response: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(jsonData)))
	w.WriteHeader(http.StatusCreated)
	w.Write(jsonData)
}
