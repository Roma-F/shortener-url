package service

import (
	"strings"
	"testing"

	"github.com/Roma-F/shortener-url/internal/app/config"
	"github.com/Roma-F/shortener-url/internal/app/storage"
	"github.com/stretchr/testify/assert"
)

func setupService() *URLService {
	cfg := &config.ServerOption{
		RunAddr:      ":8080",
		ShortUrlAddr: "http://localhost:8080",
	}
	repo := storage.NewMemoryStorage()
	return NewURLService(repo, cfg)
}

func TestURLService_FetchOriginalURL(t *testing.T) {
	svc := setupService()

	originalURL := "https://example.com"
	shortURL := svc.GenerateShortURL(originalURL)
	parts := strings.Split(shortURL, "/")
	id := parts[len(parts)-1]

	fetchedURL, err := svc.FetchOriginalURL(id)
	assert.NoError(t, err)
	assert.Equal(t, originalURL, fetchedURL)

	_, err = svc.FetchOriginalURL("nonexist")
	assert.Error(t, err)
}

func TestURLService_GenerateShortURL(t *testing.T) {
	svc := setupService()

	originalURL := "https://example.com"
	host := "localhost:8080"
	shortURL := svc.GenerateShortURL(originalURL)

	expectedPrefix := "http://" + host + "/"
	assert.True(t, strings.HasPrefix(shortURL, expectedPrefix), "short URL should start with %s", expectedPrefix)

	parts := strings.Split(shortURL, "/")
	id := parts[len(parts)-1]
	assert.Equal(t, 8, len(id))
}
