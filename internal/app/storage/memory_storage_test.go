package storage_test

import (
	"testing"

	"github.com/Roma-F/shortener-url/internal/app/storage"
	"github.com/stretchr/testify/assert"
)

func TestMemoryStorage_SaveAndFetch(t *testing.T) {
	ms := storage.NewMemoryStorage()

	id := "abc123"
	originalURL := "http://example.com"

	err := ms.Save(id, originalURL)
	assert.NoError(t, err)

	fetchedURL, err := ms.Fetch(id)
	assert.NoError(t, err)
	assert.Equal(t, originalURL, fetchedURL)
}

func TestMemoryStorage_FetchNonexistent(t *testing.T) {
	ms := storage.NewMemoryStorage()
	fetchedURL, err := ms.Fetch("nonexistent")
	assert.Error(t, err)
	assert.Empty(t, fetchedURL)
	expectedErrorMsg := "this id not found"
	assert.Equal(t, expectedErrorMsg, err.Error())
}
