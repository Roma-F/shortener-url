package service

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"

	"github.com/Roma-F/shortener-url/internal/app/storage"
)

type URLService struct {
	repo storage.Repository
}

func NewURLService(repo storage.Repository) *URLService {
	return &URLService{repo: repo}
}

func (s *URLService) FetchOriginalURL(id string) (string, error) {
	return s.repo.Fetch(id)
}

func (s *URLService) GenerateShortURL(originalURL string, host string) string {
	hash := md5.Sum([]byte(originalURL))
	id := hex.EncodeToString(hash[:])[:8]

	s.repo.Save(id, originalURL)

	return fmt.Sprintf("http://%s/%s", host, id)
}
