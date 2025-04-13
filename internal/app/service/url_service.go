package service

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"

	"github.com/Roma-F/shortener-url/internal/app/config"
)

type Repository interface {
	Save(id string, url string) error
	Fetch(id string) (string, error)
	FindByURL(url string) (string, bool)
}

type URLService struct {
	repo Repository
	cfg  *config.ServerOption
}

func NewURLService(repo Repository, cfg *config.ServerOption) *URLService {
	return &URLService{repo: repo, cfg: cfg}
}

func (s *URLService) FetchOriginalURL(id string) (string, error) {
	return s.repo.Fetch(id)
}

func (s *URLService) GenerateShortURL(originalURL string) (string, error) {
	if id, found := s.repo.FindByURL(originalURL); found {
		return fmt.Sprintf("%s/%s", s.cfg.ShortURLAddr, id), nil
	}

	hash := md5.Sum([]byte(originalURL))
	id := hex.EncodeToString(hash[:])[:8]

	if _, err := s.repo.Fetch(id); err == nil {
		unique := false
		for i := 1; i <= s.cfg.MaxAttempts; i++ {
			salt := fmt.Sprintf("%d", i)
			newHash := md5.Sum([]byte(originalURL + salt))
			newID := hex.EncodeToString(newHash[:])[:8]

			if _, err := s.repo.Fetch(newID); err != nil {
				id = newID
				unique = true
				break
			}
		}
		if !unique {
			return "", fmt.Errorf("failed to generate unique short URL for %s", originalURL)
		}
	}

	if err := s.repo.Save(id, originalURL); err != nil {
		return "", fmt.Errorf("failed to save short url: %v", err)
	}

	return fmt.Sprintf("%s/%s", s.cfg.ShortURLAddr, id), nil
}
