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
	defer m.mu.Unlock()
	m.urls[id] = url

	return nil
}

func (m *MemoryStorage) Fetch(id string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	value, ok := m.urls[id]

	if !ok {
		return "", errors.New("this id not found")
	}

	return value, nil
}

func (m *MemoryStorage) FindByURL(url string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for id, storedURL := range m.urls {
		if storedURL == url {
			return id, true
		}
	}
	return "", false
}
