package service

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/Roma-F/shortener-url/internal/app/config"
	"github.com/Roma-F/shortener-url/internal/app/models"
	"github.com/Roma-F/shortener-url/internal/app/repository"
)

type URLService struct {
	repo repository.Repository
	cfg  *config.ServerOption
}

func (s *URLService) ShortenBatch(requests []models.ShortenBatchItem) ([]models.ShortenedURLItem, error) {
	if len(requests) == 0 {
		return nil, fmt.Errorf("empty batch")
	}

	pairs := make([]models.URLPair, 0, len(requests))
	for _, req := range requests {
		hash := md5.Sum([]byte(req.OriginalURL))
		id := hex.EncodeToString(hash[:])[:8]

		if existingID, found := s.repo.FindByURL(req.OriginalURL); found {
			pairs = append(pairs, models.URLPair{
				OriginalURL:   req.OriginalURL,
				ShortURL:      existingID,
				CorrelationID: req.CorrelationID,
			})
			continue
		}

		if _, err := s.repo.Fetch(id); err == nil {
			unique := false
			for i := 1; i <= s.cfg.MaxAttempts; i++ {
				salt := fmt.Sprintf("%d", i)
				newHash := md5.Sum([]byte(req.OriginalURL + salt))
				newID := hex.EncodeToString(newHash[:])[:8]

				if _, err := s.repo.Fetch(newID); err != nil {
					id = newID
					unique = true
					break
				}
			}
			if !unique {
				return nil, fmt.Errorf("failed to generate unique short URL for %s", req.OriginalURL)
			}
		}

		pairs = append(pairs, models.URLPair{
			OriginalURL:   req.OriginalURL,
			ShortURL:      id,
			CorrelationID: req.CorrelationID,
		})
	}

	savedPairs, err := s.repo.SaveBatch(pairs)
	if err != nil {
		return nil, fmt.Errorf("failed to save batch: %v", err)
	}

	result := make([]models.ShortenedURLItem, len(savedPairs))
	for i, pair := range savedPairs {
		result[i] = models.ShortenedURLItem{
			CorrelationID: pair.CorrelationID,
			ShortURL:      fmt.Sprintf("%s/%s", s.cfg.ShortURLAddr, pair.ShortURL),
		}
	}

	return result, nil
}

func (s *URLService) ShortenBatchWithUser(requests []models.ShortenBatchItem, userID string) ([]models.ShortenedURLItem, error) {
	if len(requests) == 0 {
		return nil, fmt.Errorf("empty batch")
	}

	pairs := make([]models.URLPair, 0, len(requests))
	for _, req := range requests {
		hash := md5.Sum([]byte(req.OriginalURL))
		id := hex.EncodeToString(hash[:])[:8]

		if existingID, found := s.repo.FindByURL(req.OriginalURL); found {
			pairs = append(pairs, models.URLPair{
				OriginalURL:   req.OriginalURL,
				ShortURL:      existingID,
				CorrelationID: req.CorrelationID,
				UserID:        userID,
			})
			continue
		}

		if _, err := s.repo.Fetch(id); err == nil {
			unique := false
			for i := 1; i <= s.cfg.MaxAttempts; i++ {
				salt := fmt.Sprintf("%d", i)
				newHash := md5.Sum([]byte(req.OriginalURL + salt))
				newID := hex.EncodeToString(newHash[:])[:8]

				if _, err := s.repo.Fetch(newID); err != nil {
					id = newID
					unique = true
					break
				}
			}
			if !unique {
				return nil, fmt.Errorf("failed to generate unique short URL for %s", req.OriginalURL)
			}
		}

		pairs = append(pairs, models.URLPair{
			OriginalURL:   req.OriginalURL,
			ShortURL:      id,
			CorrelationID: req.CorrelationID,
			UserID:        userID,
		})
	}

	savedPairs, err := s.repo.SaveBatchWithUser(pairs)
	if err != nil {
		return nil, fmt.Errorf("failed to save batch: %v", err)
	}

	result := make([]models.ShortenedURLItem, len(savedPairs))
	for i, pair := range savedPairs {
		result[i] = models.ShortenedURLItem{
			CorrelationID: pair.CorrelationID,
			ShortURL:      fmt.Sprintf("%s/%s", s.cfg.ShortURLAddr, pair.ShortURL),
		}
	}

	return result, nil
}

func NewURLService(repo repository.Repository, cfg *config.ServerOption) *URLService {
	return &URLService{repo: repo, cfg: cfg}
}

func (s *URLService) FetchOriginalURL(id string) (string, error) {
	return s.repo.Fetch(id)
}

func (s *URLService) GenerateShortURL(originalURL string) (string, error) {
	if id, found := s.repo.FindByURL(originalURL); found {
		return fmt.Sprintf("%s/%s", s.cfg.ShortURLAddr, id), repository.ErrURLConflict
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
		if errors.Is(err, repository.ErrURLConflict) {
			existingID, found := s.repo.FindByURL(originalURL)
			if found {
				return fmt.Sprintf("%s/%s", s.cfg.ShortURLAddr, existingID), repository.ErrURLConflict
			}
		}
		return "", fmt.Errorf("failed to save short url: %v", err)
	}

	return fmt.Sprintf("%s/%s", s.cfg.ShortURLAddr, id), nil
}

func (s *URLService) GenerateShortURLWithUser(originalURL string, userID string) (string, error) {
	if id, found := s.repo.FindByURL(originalURL); found {
		return fmt.Sprintf("%s/%s", s.cfg.ShortURLAddr, id), repository.ErrURLConflict
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

	if err := s.repo.SaveWithUser(id, originalURL, userID); err != nil {
		if errors.Is(err, repository.ErrURLConflict) {
			existingID, found := s.repo.FindByURL(originalURL)
			if found {
				return fmt.Sprintf("%s/%s", s.cfg.ShortURLAddr, existingID), repository.ErrURLConflict
			}
		}
		return "", fmt.Errorf("failed to save short url: %v", err)
	}

	return fmt.Sprintf("%s/%s", s.cfg.ShortURLAddr, id), nil
}

func (s *URLService) GetUserURLs(userID string) ([]models.UserURL, error) {
	urls, err := s.repo.GetUserURLs(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user URLs: %v", err)
	}

	for i := range urls {
		urls[i].ShortURL = fmt.Sprintf("%s/%s", s.cfg.ShortURLAddr, urls[i].ShortURL)
	}

	return urls, nil
}
