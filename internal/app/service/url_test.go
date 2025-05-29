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
		MaxAttempts:  10,
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
	shortURL, err := svc.GenerateShortURL(originalURL)
	assert.NoError(t, err)

	expectedPrefix := "http://localhost:8080/"
	assert.True(t, strings.HasPrefix(shortURL, expectedPrefix), "short URL should start with %s", expectedPrefix)

	parts := strings.Split(shortURL, "/")
	id := parts[len(parts)-1]
	assert.Equal(t, 8, len(id))
}

func TestURLService_GenerateShortURLWithUser(t *testing.T) {
	svc := setupService()

	originalURL := "https://example.com"
	userID := "test-user-123"

	shortURL, err := svc.GenerateShortURLWithUser(originalURL, userID)
	assert.NoError(t, err)

	expectedPrefix := "http://localhost:8080/"
	assert.True(t, strings.HasPrefix(shortURL, expectedPrefix), "short URL should start with %s", expectedPrefix)

	parts := strings.Split(shortURL, "/")
	id := parts[len(parts)-1]
	assert.Equal(t, 8, len(id))

	fetchedURL, err := svc.FetchOriginalURL(id)
	assert.NoError(t, err)
	assert.Equal(t, originalURL, fetchedURL)
}

func TestURLService_GetUserURLs(t *testing.T) {
	svc := setupService()

	userID := "test-user-123"

	_, err := svc.GenerateShortURLWithUser("https://example1.com", userID)
	assert.NoError(t, err)
	_, err = svc.GenerateShortURLWithUser("https://example2.com", userID)
	assert.NoError(t, err)

	_, err = svc.GenerateShortURLWithUser("https://example3.com", "other-user")
	assert.NoError(t, err)

	urls, err := svc.GetUserURLs(userID)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(urls))

	originalURLs := make(map[string]bool)
	for _, url := range urls {
		originalURLs[url.OriginalURL] = true
		assert.True(t, strings.HasPrefix(url.ShortURL, "http://localhost:8080/"))
	}

	assert.True(t, originalURLs["https://example1.com"])
	assert.True(t, originalURLs["https://example2.com"])
	assert.False(t, originalURLs["https://example3.com"])
}

func TestURLService_GetUserURLs_EmptyResult(t *testing.T) {
	svc := setupService()

	urls, err := svc.GetUserURLs("nonexistent-user")
	assert.NoError(t, err)
	assert.Equal(t, 0, len(urls))
}

func TestURLService_ShortenBatch(t *testing.T) {
	svc := setupService()

	requests := []models.ShortenBatchItem{
		{
			CorrelationID: "1",
			OriginalURL:   "https://example1.com",
		},
		{
			CorrelationID: "2",
			OriginalURL:   "https://example2.com",
		},
	}

	result, err := svc.ShortenBatch(requests)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(result))

	assert.Equal(t, "1", result[0].CorrelationID)
	assert.Equal(t, "2", result[1].CorrelationID)

	expectedPrefix := "http://localhost:8080/"
	assert.True(t, strings.HasPrefix(result[0].ShortURL, expectedPrefix))
	assert.True(t, strings.HasPrefix(result[1].ShortURL, expectedPrefix))

	parts1 := strings.Split(result[0].ShortURL, "/")
	id1 := parts1[len(parts1)-1]
	url1, err := svc.FetchOriginalURL(id1)
	assert.NoError(t, err)
	assert.Equal(t, "https://example1.com", url1)

	parts2 := strings.Split(result[1].ShortURL, "/")
	id2 := parts2[len(parts2)-1]
	url2, err := svc.FetchOriginalURL(id2)
	assert.NoError(t, err)
	assert.Equal(t, "https://example2.com", url2)
}

func TestURLService_ShortenBatchWithUser(t *testing.T) {
	svc := setupService()

	userID := "test-user-123"
	requests := []models.ShortenBatchItem{
		{
			CorrelationID: "1",
			OriginalURL:   "https://example1.com",
		},
		{
			CorrelationID: "2",
			OriginalURL:   "https://example2.com",
		},
	}

	result, err := svc.ShortenBatchWithUser(requests, userID)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(result))

	assert.Equal(t, "1", result[0].CorrelationID)
	assert.Equal(t, "2", result[1].CorrelationID)

	expectedPrefix := "http://localhost:8080/"
	assert.True(t, strings.HasPrefix(result[0].ShortURL, expectedPrefix))
	assert.True(t, strings.HasPrefix(result[1].ShortURL, expectedPrefix))

	urls, err := svc.GetUserURLs(userID)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(urls))
}

func TestURLService_ShortenBatch_EmptyBatch(t *testing.T) {
	svc := setupService()

	requests := []models.ShortenBatchItem{}
	result, err := svc.ShortenBatch(requests)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestURLService_ShortenBatchWithUser_EmptyBatch(t *testing.T) {
	svc := setupService()

	requests := []models.ShortenBatchItem{}
	result, err := svc.ShortenBatchWithUser(requests, "test-user")
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestURLService_ShortenBatch_DuplicateURLs(t *testing.T) {
	svc := setupService()

	requests := []models.ShortenBatchItem{
		{
			CorrelationID: "1",
			OriginalURL:   "https://example.com",
		},
		{
			CorrelationID: "2",
			OriginalURL:   "https://example.com",
		},
	}

	result, err := svc.ShortenBatch(requests)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(result))

	assert.Equal(t, result[0].ShortURL, result[1].ShortURL)
}

func TestURLService_UserIsolation(t *testing.T) {
	svc := setupService()

	userID1 := "user-1"
	userID2 := "user-2"

	_, err := svc.GenerateShortURLWithUser("https://example1.com", userID1)
	assert.NoError(t, err)
	_, err = svc.GenerateShortURLWithUser("https://example2.com", userID1)
	assert.NoError(t, err)

	_, err = svc.GenerateShortURLWithUser("https://example3.com", userID2)
	assert.NoError(t, err)

	urls1, err := svc.GetUserURLs(userID1)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(urls1))

	urls2, err := svc.GetUserURLs(userID2)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(urls2))

	user1URLs := make(map[string]bool)
	for _, url := range urls1 {
		user1URLs[url.OriginalURL] = true
	}

	for _, url := range urls2 {
		assert.False(t, user1URLs[url.OriginalURL], "URL should not belong to user1")
	}
}
