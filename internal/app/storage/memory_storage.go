package storage

import "errors"

type Repository interface {
	Save(id string, url string) error
	Fetch(id string) (string, error)
}

type MemoryStorage struct {
	urls map[string]string
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{urls: make(map[string]string)}
}

func (m *MemoryStorage) Save(id string, url string) error {
	m.urls[id] = url
	return nil
}

func (m *MemoryStorage) Fetch(id string) (string, error) {
	_, ok := m.urls[id]
	if !ok {
		return "", errors.New("this id not found")
	}

	return m.urls[id], nil
}
