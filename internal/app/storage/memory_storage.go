package storage

import (
	"errors"
	"sync"
)

type MemoryStorage struct {
	mu   sync.RWMutex
	urls map[string]string
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{urls: make(map[string]string)}
}

func (m *MemoryStorage) Save(id string, url string) error {
	m.mu.Lock()
	m.urls[id] = url
	m.mu.Unlock()

	return nil

}

func (m *MemoryStorage) Fetch(id string) (string, error) {
	m.mu.RLock()
	value, ok := m.urls[id]
	m.mu.RUnlock()

	if !ok {
		return "", errors.New("this id not found")
	}

	return value, nil
}
