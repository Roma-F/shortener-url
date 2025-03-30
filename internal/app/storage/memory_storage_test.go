package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	tId         = "abc123"
	originalURL = "http://example.com"
)

func TestMemoryStorage_Save(t *testing.T) {
	ms := NewMemoryStorage()

	err := ms.Save(tId, originalURL)
	assert.NoError(t, err)

	ms.mu.RLock()
	storedURL, ok := ms.urls[tId]
	ms.mu.RUnlock()

	assert.True(t, ok)
	assert.Equal(t, originalURL, storedURL)
}

func TestMemoryStorage_Fetch(t *testing.T) {
	ms := NewMemoryStorage()

	ms.mu.Lock()
	ms.urls[tId] = originalURL
	ms.mu.Unlock()

	fetchedURL, err := ms.Fetch(tId)
	assert.NoError(t, err)
	assert.Equal(t, originalURL, fetchedURL)
}

func TestMemoryStorage_FetchNonexistent(t *testing.T) {
	ms := NewMemoryStorage()

	fetchedURL, err := ms.Fetch("nonexistent")
	assert.Error(t, err)
	assert.Empty(t, fetchedURL)
	expectedErrorMsg := "this id not found"
	assert.Equal(t, expectedErrorMsg, err.Error())
}
