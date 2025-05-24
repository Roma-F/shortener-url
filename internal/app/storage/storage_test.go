package storage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Roma-F/shortener-url/internal/app/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	tID         = "abc123"
	originalURL = "http://example.com"
)

func TestMemoryStorage_Save(t *testing.T) {
	ms := NewMemoryStorage("")

	err := ms.Save(tID, originalURL)
	assert.NoError(t, err)

	ms.mu.RLock()
	record, ok := ms.records[tID]
	ms.mu.RUnlock()

	assert.True(t, ok)
	assert.Equal(t, originalURL, record.OriginalURL)
	assert.Equal(t, 1, record.ID)
}

func TestMemoryStorage_Fetch(t *testing.T) {
	ms := NewMemoryStorage("")

	ms.mu.Lock()
	ms.records[tID] = URLRecord{
		ID:          1,
		ShortURL:    tID,
		OriginalURL: originalURL,
	}
	ms.mu.Unlock()

	fetchedURL, err := ms.Fetch(tID)
	assert.NoError(t, err)
	assert.Equal(t, originalURL, fetchedURL)
}

func TestMemoryStorage_FetchNonexistent(t *testing.T) {
	ms := NewMemoryStorage("")

	fetchedURL, err := ms.Fetch("nonexistent")
	assert.Error(t, err)
	assert.Empty(t, fetchedURL)
	expectedErrorMsg := "short URL not found"
	assert.Equal(t, expectedErrorMsg, err.Error())
}

func TestMemoryStorage_FindByURL(t *testing.T) {
	ms := NewMemoryStorage("")

	ms.mu.Lock()
	ms.records[tID] = URLRecord{
		ID:          1,
		ShortURL:    tID,
		OriginalURL: originalURL,
	}
	ms.mu.Unlock()

	foundID, found := ms.FindByURL(originalURL)
	assert.True(t, found)
	assert.Equal(t, tID, foundID)

	foundID, found = ms.FindByURL("http://nonexistent.com")
	assert.False(t, found)
	assert.Empty(t, foundID)
}

func TestMemoryStorage_SaveToFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "url_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	tempFile := filepath.Join(tempDir, "urls.json")

	ms := NewMemoryStorage(tempFile)

	err = ms.Save("short1", "http://example1.com")
	require.NoError(t, err)
	err = ms.Save("short2", "http://example2.com")
	require.NoError(t, err)

	_, err = os.Stat(tempFile)
	assert.NoError(t, err)

	content, err := os.ReadFile(tempFile)
	require.NoError(t, err)

	contentStr := string(content)
	assert.Contains(t, contentStr, "http://example1.com")
	assert.Contains(t, contentStr, "http://example2.com")
	assert.Contains(t, contentStr, "short1")
	assert.Contains(t, contentStr, "short2")
}

func TestMemoryStorage_LoadFromFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "url_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	tempFile := filepath.Join(tempDir, "urls.json")

	testData := `{"id":1,"short_url":"short1","original_url":"http://example1.com"}
{"id":2,"short_url":"short2","original_url":"http://example2.com"}`
	err = os.WriteFile(tempFile, []byte(testData), 0644)
	require.NoError(t, err)

	ms := NewMemoryStorage(tempFile)

	assert.Equal(t, 2, len(ms.records))

	url1, err := ms.Fetch("short1")
	assert.NoError(t, err)
	assert.Equal(t, "http://example1.com", url1)

	url2, err := ms.Fetch("short2")
	assert.NoError(t, err)
	assert.Equal(t, "http://example2.com", url2)

	assert.Equal(t, 2, ms.lastID)
}

func TestMemoryStorage_FilePathEmpty(t *testing.T) {
	ms := NewMemoryStorage("")

	err := ms.Save("short1", "http://example1.com")
	assert.NoError(t, err)

	url, err := ms.Fetch("short1")
	assert.NoError(t, err)
	assert.Equal(t, "http://example1.com", url)
}

func TestMemoryStorage_FilePersistence(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "url_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	tempFile := filepath.Join(tempDir, "urls.json")

	ms1 := NewMemoryStorage(tempFile)
	err = ms1.Save("short1", "http://example1.com")
	require.NoError(t, err)

	ms2 := NewMemoryStorage(tempFile)

	url, err := ms2.Fetch("short1")
	assert.NoError(t, err)
	assert.Equal(t, "http://example1.com", url)
}

func TestMemoryStorage_SaveBatch(t *testing.T) {
	ms := NewMemoryStorage("")

	pairs := []models.URLPair{
		{
			OriginalURL:   "http://example1.com",
			ShortURL:      "short1",
			CorrelationID: "1",
		},
		{
			OriginalURL:   "http://example2.com",
			ShortURL:      "short2",
			CorrelationID: "2",
		},
	}

	savedPairs, err := ms.SaveBatch(pairs)
	assert.NoError(t, err)
	assert.Equal(t, pairs, savedPairs)

	url1, err := ms.Fetch("short1")
	assert.NoError(t, err)
	assert.Equal(t, "http://example1.com", url1)

	url2, err := ms.Fetch("short2")
	assert.NoError(t, err)
	assert.Equal(t, "http://example2.com", url2)
}

func TestMemoryStorage_SaveBatch_WithFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "url_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	tempFile := filepath.Join(tempDir, "urls.json")
	ms := NewMemoryStorage(tempFile)

	pairs := []models.URLPair{
		{
			OriginalURL:   "http://example1.com",
			ShortURL:      "short1",
			CorrelationID: "1",
		},
		{
			OriginalURL:   "http://example2.com",
			ShortURL:      "short2",
			CorrelationID: "2",
		},
	}

	savedPairs, err := ms.SaveBatch(pairs)
	assert.NoError(t, err)
	assert.Equal(t, pairs, savedPairs)

	content, err := os.ReadFile(tempFile)
	require.NoError(t, err)

	contentStr := string(content)
	assert.Contains(t, contentStr, "http://example1.com")
	assert.Contains(t, contentStr, "http://example2.com")
	assert.Contains(t, contentStr, "short1")
	assert.Contains(t, contentStr, "short2")
}
