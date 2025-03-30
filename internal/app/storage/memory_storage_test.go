package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	tID         = "abc123"
	originalURL = "http://example.com"
)

func TestMemoryStorage_Save(t *testing.T) {
	ms := NewMemoryStorage()

	err := ms.Save(tID, originalURL)
	assert.NoError(t, err)

	ms.mu.RLock()
	storedURL, ok := ms.urls[tID]
	ms.mu.RUnlock()

	assert.True(t, ok)
	assert.Equal(t, originalURL, storedURL)
}

func TestMemoryStorage_Fetch(t *testing.T) {
	ms := NewMemoryStorage()

	ms.mu.Lock()
	ms.urls[tID] = originalURL
	ms.mu.Unlock()

	fetchedURL, err := ms.Fetch(tID)
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
