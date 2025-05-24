package repository

import (
	"errors"

	"github.com/Roma-F/shortener-url/internal/app/models"
)

var ErrURLConflict = errors.New("original URL already exists")

type Repository interface {
	Save(id string, url string) error
	Fetch(id string) (string, error)
	FindByURL(url string) (string, bool)
	SaveBatch(pairs []models.URLPair) ([]models.URLPair, error)
}
