package storage

import (
	"github.com/Roma-F/shortener-url/internal/app/config"
	"github.com/Roma-F/shortener-url/internal/app/logger"
	"github.com/Roma-F/shortener-url/internal/app/repository"
)

func NewRepository(cfg *config.ServerOption) (repository.Repository, *PostgresStorage) {
	var repo repository.Repository
	var dbStorage *PostgresStorage
	var err error

	if cfg.DatabaseDSN != "" {
		dbStorage, err = NewPostgresStorage(cfg)
		if err == nil {
			logger.Sugar.Infow("Using PostgreSQL database for URL storage")
			repo = dbStorage
		} else {
			logger.Sugar.Warnw("Failed to connect to PostgreSQL database, falling back to file storage", "error", err)
		}
	}

	if repo == nil && cfg.FSPath != "" {
		logger.Sugar.Infow("Using file storage for URL storage", "path", cfg.FSPath)
		repo = NewMemoryStorage(cfg.FSPath)
	}

	if repo == nil {
		logger.Sugar.Infow("Using in-memory storage for URL storage")
		repo = NewMemoryStorage("")
	}

	return repo, dbStorage
}
