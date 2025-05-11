package storage

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Roma-F/shortener-url/internal/app/config"
	"github.com/Roma-F/shortener-url/internal/app/logger"
	"github.com/Roma-F/shortener-url/internal/app/models"
	"github.com/Roma-F/shortener-url/internal/app/repository"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

//go:embed ddl.sql
var queries string
var queryMap map[string]string

type PostgresStorage struct {
	db *sqlx.DB
}

func init() {
	queryMap = parseQueries(queries)
}

func (p *PostgresStorage) SaveBatch(pairs []models.URLPair) ([]models.URLPair, error) {
	tx, err := p.db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}

	defer func() {
		if err := tx.Rollback(); err != nil {
			logger.Sugar.Errorw("Failed to rollback transaction", "error", err)
		}
	}()

	stmt, err := tx.Preparex(getQuery("save-url"))
	if err != nil {
		return nil, fmt.Errorf("Failed to prepare statement: %w", err)
	}

	defer func() {
		if err := stmt.Close(); err != nil {
			logger.Sugar.Errorw("Failed close statement", "error", err)
		}
	}()

	for _, pair := range pairs {
		_, err := stmt.Exec(pair.ShortURL, pair.OriginalURL)
		if err != nil {
			return nil, fmt.Errorf("failed to save URL pair: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return pairs, nil
}

func (p *PostgresStorage) Fetch(shortURL string) (string, error) {
	var originalURL string

	err := p.db.QueryRow(getQuery("fetch-url"), shortURL).Scan(&originalURL)
	if err != nil {
		return "", fmt.Errorf("short URL not found: %w", err)
	}

	return originalURL, nil
}

func (p *PostgresStorage) FindByURL(originalURL string) (string, bool) {
	var shortURL string

	err := p.db.QueryRow(getQuery("find-by-original-url"), originalURL).Scan(&shortURL)
	if err != nil {
		return "", false
	}

	return shortURL, true
}

func (p *PostgresStorage) Save(shortURL string, originalURL string) error {
	_, err := p.db.Exec(getQuery("save-url-check-conflict"), shortURL, originalURL)
	if err != nil {
		return fmt.Errorf("error saving URL: %w", err)
	}
	var count int
	err = p.db.QueryRow(getQuery("check-short-url-exists"), shortURL).Scan(&count)
	if err != nil {
		return fmt.Errorf("error checking URL record: %w", err)
	}

	if count == 0 {
		return fmt.Errorf("duplicate original URL: %w", repository.ErrURLConflict)
	}

	return nil
}

func parseQueries(queriesText string) map[string]string {
	queries := make(map[string]string)

	lines := strings.Split(queriesText, "\n")
	var currentQuery strings.Builder
	var currentName string

	for _, line := range lines {
		if strings.HasPrefix(line, "-- name:") {
			if currentName != "" && currentQuery.Len() > 0 {
				queries[currentName] = strings.TrimSpace(currentQuery.String())
				currentQuery.Reset()
			}

			currentName = strings.TrimSpace(strings.TrimPrefix(line, "-- name:"))
		} else if currentName != "" && !strings.HasPrefix(line, "--") {
			currentQuery.WriteString(line)
			currentQuery.WriteString("\n")
		}
	}

	if currentName != "" && currentQuery.Len() > 0 {
		queries[currentName] = strings.TrimSpace(currentQuery.String())
	}

	return queries
}

func getQuery(name string) string {
	query, ok := queryMap[name]
	if !ok {
		panic(fmt.Sprintf("SQL query with name '%s' not found", name))
	}
	return query
}

func NewPostgresStorage(cfg *config.ServerOption) (*PostgresStorage, error) {
	if cfg.DatabaseDSN == "" {
		return nil, fmt.Errorf("database DSN is empty")
	}

	db, err := sqlx.Connect("postgres", cfg.DatabaseDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if cfg.ApplyMigrations {
		if err := runMigrationsWithTool(cfg); err != nil {
			logger.Sugar.Warnw("Failed to apply migrations", "error", err)
		}
	}

	return &PostgresStorage{db: db}, nil
}

func runMigrationsWithTool(cfg *config.ServerOption) error {
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	migrationsPath := filepath.Join(currentDir, cfg.MigrationsPath)

	_, err = os.Stat(migrationsPath)
	if os.IsNotExist(err) {
		logger.Sugar.Infow("Migrations directory not found, skipping migrations", "path", migrationsPath)
		return nil
	}

	logger.Sugar.Infow("Using migrations directory", "path", migrationsPath)

	cmd := exec.Command(
		"go", "run", filepath.Join(currentDir, "cmd/migrator/main.go"),
		"-d", cfg.DatabaseDSN,
		"-p", migrationsPath,
		"-t", cfg.MigrationsTable,
	)

	logger.Sugar.Debugw("Running migration command", "command", cmd.String())

	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	if err != nil {
		if strings.Contains(outputStr, "no migrations to apply") {
			logger.Sugar.Info("No migrations to apply")
			return nil
		}
		return fmt.Errorf("migration tool failed: %s: %w", outputStr, err)
	}

	logger.Sugar.Infow("Migrations applied successfully", "output", outputStr)
	return nil
}

func (p *PostgresStorage) Ping(ctx context.Context) error {
	return p.db.PingContext(ctx)
}

func (p *PostgresStorage) Close() error {
	return p.db.Close()
}
