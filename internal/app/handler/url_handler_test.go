package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Roma-F/shortener-url/internal/app/service"
	"github.com/Roma-F/shortener-url/internal/app/storage"
	"github.com/stretchr/testify/assert"
)

func setupHandler() *URLHandler {
	repo := storage.NewMemoryStorage()
	svc := service.NewURLService(repo)
	return NewURLHandler(svc)
}

func TestURLHandler_ShortenURL_Success(t *testing.T) {
	handler := setupHandler()

	originalURL := "https://example.com"
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(originalURL))
	req.Header.Set("Content-Type", "text/plain")
	req.Host = "example.com"

	rr := httptest.NewRecorder()
	handler.ShortenURL(rr, req)

	resp := rr.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	shortURL := string(body)

	expectedPrefix := "http://" + req.Host + "/"
	assert.True(t, strings.HasPrefix(shortURL, expectedPrefix), "short URL should start with %s", expectedPrefix)

	parts := strings.Split(shortURL, "/")
	id := parts[len(parts)-1]
	assert.Equal(t, 8, len(id))

	getReq := httptest.NewRequest(http.MethodGet, "/"+id, nil)
	getRespRec := httptest.NewRecorder()
	handler.GetMainURL(getRespRec, getReq)

	getResp := getRespRec.Result()
	defer getResp.Body.Close()
}

func TestURLHandler_ShortenURL_MethodNotAllowed(t *testing.T) {
	handler := setupHandler()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Content-Type", "text/plain")
	req.Host = "example.com"
	rr := httptest.NewRecorder()

	handler.ShortenURL(rr, req)

	resp := rr.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
}

func TestURLHandler_ShortenURL_InvalidContentType(t *testing.T) {
	handler := setupHandler()

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://example.com"))
	req.Header.Set("Content-Type", "application/json")
	req.Host = "example.com"
	rr := httptest.NewRecorder()

	handler.ShortenURL(rr, req)

	resp := rr.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestURLHandler_GetMainURL_MethodNotAllowed(t *testing.T) {
	handler := setupHandler()

	req := httptest.NewRequest(http.MethodPost, "/someid", nil)
	rr := httptest.NewRecorder()

	handler.GetMainURL(rr, req)

	resp := rr.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
}

func TestURLHandler_GetMainURL_NotFound(t *testing.T) {
	handler := setupHandler()

	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	rr := httptest.NewRecorder()

	handler.GetMainURL(rr, req)

	resp := rr.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestURLHandler_GetMainURL_Success(t *testing.T) {
	handler := setupHandler()

	originalURL := "https://example.com"
	reqShorten := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(originalURL))
	reqShorten.Header.Set("Content-Type", "text/plain")
	reqShorten.Host = "example.com"
	rrShorten := httptest.NewRecorder()
	handler.ShortenURL(rrShorten, reqShorten)
	respShorten := rrShorten.Result()
	defer respShorten.Body.Close()
	body, err := io.ReadAll(respShorten.Body)
	assert.NoError(t, err)
	shortURL := string(body)
	parts := strings.Split(shortURL, "/")
	id := parts[len(parts)-1]

	reqGet := httptest.NewRequest(http.MethodGet, "/"+id, nil)
	rrGet := httptest.NewRecorder()
	handler.GetMainURL(rrGet, reqGet)
	respGet := rrGet.Result()
	defer respGet.Body.Close()
}
