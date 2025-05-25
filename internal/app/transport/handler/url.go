package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/Roma-F/shortener-url/internal/app/auth"
	"github.com/Roma-F/shortener-url/internal/app/logger"
	"github.com/Roma-F/shortener-url/internal/app/models"
	"github.com/Roma-F/shortener-url/internal/app/repository"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type URLShortener interface {
	FetchOriginalURL(id string) (string, error)
	GenerateShortURL(originalURL string) (string, error)
	GenerateShortURLWithUser(originalURL string, userID string) (string, error)
	ShortenBatch(requests []models.ShortenBatchItem) ([]models.ShortenedURLItem, error)
	ShortenBatchWithUser(requests []models.ShortenBatchItem, userID string) ([]models.ShortenedURLItem, error)
	GetUserURLs(userID string) ([]models.UserURL, error)
}

type DeleteService interface {
	AddDeleteTasks(userID string, shortURLs []string)
}

type URLHandler struct {
	service       URLShortener
	deleteService DeleteService
}

func NewURLHandler(svc URLShortener, deleteService DeleteService) *URLHandler {
	return &URLHandler{
		service:       svc,
		deleteService: deleteService,
	}
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
	userID := auth.GetUserIDFromContext(r.Context())

	var shortURL string
	if userID != "" {
		shortURL, err = h.service.GenerateShortURLWithUser(url, userID)
	} else {
		shortURL, err = h.service.GenerateShortURL(url)
	}

	if err != nil {
		if errors.Is(err, repository.ErrURLConflict) {
			w.Header().Set("Content-Type", "text/plain")
			w.Header().Set("Content-Length", strconv.Itoa(len(shortURL)))
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(shortURL))
			return
		}
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

	userID := auth.GetUserIDFromContext(r.Context())

	var shortURL string
	var err error
	if userID != "" {
		shortURL, err = h.service.GenerateShortURLWithUser(req.URL, userID)
	} else {
		shortURL, err = h.service.GenerateShortURL(req.URL)
	}

	if err != nil {
		if errors.Is(err, repository.ErrURLConflict) {
			resp := models.ShortenURLResp{
				Result: shortURL,
			}

			jsonData, jsonErr := json.MarshalIndent(resp, "", "   ")
			if jsonErr != nil {
				http.Error(w, "Error creating JSON response: "+jsonErr.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Content-Length", strconv.Itoa(len(jsonData)))
			w.WriteHeader(http.StatusConflict)
			w.Write(jsonData)
			return
		}
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
		if errors.Is(err, repository.ErrURLDeleted) {
			w.WriteHeader(http.StatusGone)
			return
		}
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

	userID := auth.GetUserIDFromContext(r.Context())

	var shortenBatch []models.ShortenedURLItem
	var err error
	if userID != "" {
		shortenBatch, err = h.service.ShortenBatchWithUser(records, userID)
	} else {
		shortenBatch, err = h.service.ShortenBatch(records)
	}

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

func (h *URLHandler) GetUserURLs(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	urls, err := h.service.GetUserURLs(userID)
	if err != nil {
		logger.Sugar.Errorw("Failed to get user URLs", "error", err, "userID", userID)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(urls) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	jsonData, err := json.MarshalIndent(urls, "", "   ")
	if err != nil {
		http.Error(w, "Error creating JSON response: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(jsonData)))
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

func (h *URLHandler) DeleteUserURLs(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	contentType := r.Header.Get("Content-Type")
	if contentType != "" && !strings.HasPrefix(contentType, "application/json") {
		http.Error(w, "Content-Type must be application/json", http.StatusBadRequest)
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

	var shortURLs models.DeleteURLsRequest
	if err := json.Unmarshal(body, &shortURLs); err != nil {
		logger.Sugar.Debug("cannot decode request JSON body", zap.Error(err))
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if len(shortURLs) == 0 {
		http.Error(w, "Empty URL list", http.StatusBadRequest)
		return
	}

	cleanShortURLs := make([]string, len(shortURLs))
	for i, url := range shortURLs {
		if strings.Contains(url, "/") {
			parts := strings.Split(url, "/")
			cleanShortURLs[i] = parts[len(parts)-1]
		} else {
			cleanShortURLs[i] = url
		}
	}

	h.deleteService.AddDeleteTasks(userID, cleanShortURLs)

	logger.Sugar.Infow("Accepted delete request",
		"userID", userID,
		"urlCount", len(cleanShortURLs))

	w.WriteHeader(http.StatusAccepted)
}
