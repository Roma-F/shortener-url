package storage

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/Roma-F/shortener-url/internal/app/logger"
	"github.com/Roma-F/shortener-url/internal/app/models"
	"github.com/Roma-F/shortener-url/internal/app/repository"
)

type URLRecord struct {
	ID          int    `json:"id"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type MemoryStorage struct {
	mu       sync.RWMutex
	filePath string
	records  map[string]URLRecord
	lastID   int
}

func NewMemoryStorage(filePath string) *MemoryStorage {
	ms := &MemoryStorage{
		records:  make(map[string]URLRecord),
		filePath: filePath,
		lastID:   0,
	}

	if err := ms.LoadFromFile(); err != nil {
		logger.Sugar.Infow("Could not load data from file", "error", err)
	}

	return ms
}

func (m *MemoryStorage) LoadFromFile() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.filePath == "" || m.filePath == "/" {
		return nil
	}
	_, err := os.Stat(m.filePath)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return fmt.Errorf("error checking file: %w", err)
	}

	file, err := os.Open(m.filePath)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}

	defer func() {
		if err := file.Close(); err != nil {
			logger.Sugar.Infow("Failed to close file during load", "error", err)
		}
	}()

	scanner := bufio.NewScanner(file)
	maxID := 0

	for scanner.Scan() {
		line := scanner.Text()
		var record URLRecord

		if err := json.Unmarshal([]byte(line), &record); err != nil {
			return fmt.Errorf("error unmarshaling record: %w", err)
		}

		m.records[record.ShortURL] = record

		if record.ID > maxID {
			maxID = record.ID
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	m.lastID = maxID
	return nil
}

func (m *MemoryStorage) SaveToFile() error {
	if m.filePath == "" || m.filePath == "/" {
		return nil
	}

	dir := filepath.Dir(m.filePath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("error creating directory: %w", err)
		}
	}

	tempFile := m.filePath + ".tmp"
	file, err := os.Create(tempFile)

	if err != nil {
		return fmt.Errorf("error creating temp file: %w", err)
	}

	defer func() {
		if file != nil {
			if err := file.Close(); err != nil {
				logger.Sugar.Infow("Failed to close temporary file", "error", err)
			}
		}
	}()

	for _, record := range m.records {
		data, err := json.Marshal(record)
		if err != nil {
			return fmt.Errorf("error marshaling record: %w", err)
		}

		if _, err := file.Write(data); err != nil {
			return fmt.Errorf("error writing to file: %w", err)
		}

		if _, err := file.Write([]byte("\n")); err != nil {
			return fmt.Errorf("error writing newline to file: %w", err)
		}
	}

	if err := file.Close(); err != nil {
		return fmt.Errorf("error closing temp file: %w", err)
	}

	if err := os.Rename(tempFile, m.filePath); err != nil {
		return fmt.Errorf("error renaming temp file: %w", err)
	}

	return nil
}

func (m *MemoryStorage) Save(shortURL string, originalURL string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, record := range m.records {
		if record.OriginalURL == originalURL {
			return repository.ErrURLConflict
		}
	}

	m.lastID++
	record := URLRecord{
		ID:          m.lastID,
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	}

	m.records[shortURL] = record

	return m.SaveToFile()
}

func (m *MemoryStorage) Fetch(shortURL string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	record, ok := m.records[shortURL]
	if !ok {
		return "", errors.New("short URL not found")
	}

	return record.OriginalURL, nil
}

func (m *MemoryStorage) FindByURL(originalURL string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for shortURL, record := range m.records {
		if record.OriginalURL == originalURL {
			return shortURL, true
		}
	}
	return "", false
}

func (m *MemoryStorage) SaveBatch(pairs []models.URLPair) ([]models.URLPair, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, pair := range pairs {
		m.lastID++
		record := URLRecord{
			ID:          m.lastID,
			ShortURL:    pair.ShortURL,
			OriginalURL: pair.OriginalURL,
		}

		m.records[pair.ShortURL] = record
	}

	if err := m.SaveToFile(); err != nil {
		return nil, fmt.Errorf("failed to save batch to file: %w", err)
	}

	return pairs, nil
}
