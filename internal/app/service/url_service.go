package service

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"

	"github.com/Roma-F/shortener-url/internal/app/config"
	"github.com/Roma-F/shortener-url/internal/app/storage"
)

type URLService struct {
	repo storage.Repository
	cfg  *config.ServerOption
}

func NewURLService(repo storage.Repository, cfg *config.ServerOption) *URLService {
	return &URLService{repo: repo, cfg: cfg}
}

func (s *URLService) FetchOriginalURL(id string) (string, error) {
	return s.repo.Fetch(id)
}

func (s *URLService) GenerateShortURL(originalURL string) string {
	hash := md5.Sum([]byte(originalURL))
	id := hex.EncodeToString(hash[:])[:8]

	s.repo.Save(id, originalURL)

	return fmt.Sprintf("%s/%s", s.cfg.ShortUrlAddr, id)
}
