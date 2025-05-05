package service

import (
	"strings"
	"testing"

	"github.com/Roma-F/shortener-url/internal/app/config"
	"github.com/Roma-F/shortener-url/internal/app/models"
	"github.com/Roma-F/shortener-url/internal/app/storage"
	"github.com/stretchr/testify/assert"
)

func setupService() *URLService {
	cfg := &config.ServerOption{
		RunAddr:      ":8080",
		ShortURLAddr: "http://localhost:8080",
	}
	repo := storage.NewMemoryStorage("")
	return NewURLService(repo, cfg)
}

func TestURLService_FetchOriginalURL(t *testing.T) {
	svc := setupService()

	originalURL := "https://example.com"
	shortURL, _ := svc.GenerateShortURL(originalURL)

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
	shortURL, _ := svc.GenerateShortURL(originalURL)

	expectedPrefix := "http://" + host + "/"
	assert.True(t, strings.HasPrefix(shortURL, expectedPrefix), "short URL should start with %s", expectedPrefix)

	parts := strings.Split(shortURL, "/")
	id := parts[len(parts)-1]
	assert.Equal(t, 8, len(id))
}

func TestURLService_ShortenBatch(t *testing.T) {
	svc := setupService()

	requests := []models.ShortenBatchItem{
		{
			CorrelationId: "1",
			OriginalUrl:   "https://example1.com",
		},
		{
			CorrelationId: "2",
			OriginalUrl:   "https://example2.com",
		},
	}

	result, err := svc.ShortenBatch(requests)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(result))

	assert.Equal(t, "1", result[0].CorrelationId)
	assert.Equal(t, "2", result[1].CorrelationId)

	expectedPrefix := "http://localhost:8080/"
	assert.True(t, strings.HasPrefix(result[0].ShortUrl, expectedPrefix))
	assert.True(t, strings.HasPrefix(result[1].ShortUrl, expectedPrefix))

	parts1 := strings.Split(result[0].ShortUrl, "/")
	id1 := parts1[len(parts1)-1]
	url1, err := svc.FetchOriginalURL(id1)
	assert.NoError(t, err)
	assert.Equal(t, "https://example1.com", url1)

	parts2 := strings.Split(result[1].ShortUrl, "/")
	id2 := parts2[len(parts2)-1]
	url2, err := svc.FetchOriginalURL(id2)
	assert.NoError(t, err)
	assert.Equal(t, "https://example2.com", url2)
}

func TestURLService_ShortenBatch_EmptyBatch(t *testing.T) {
	svc := setupService()

	requests := []models.ShortenBatchItem{}
	result, err := svc.ShortenBatch(requests)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestURLService_ShortenBatch_DuplicateURLs(t *testing.T) {
	svc := setupService()

	requests := []models.ShortenBatchItem{
		{
			CorrelationId: "1",
			OriginalUrl:   "https://example.com",
		},
		{
			CorrelationId: "2",
			OriginalUrl:   "https://example.com",
		},
	}

	result, err := svc.ShortenBatch(requests)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(result))

	assert.Equal(t, result[0].ShortUrl, result[1].ShortUrl)
}
