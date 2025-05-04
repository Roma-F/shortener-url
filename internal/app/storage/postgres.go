package storage

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Roma-F/shortener-url/internal/app/logger"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

//go:embed ddl.sql
var queries string
var queryMap map[string]string

type PostgresStorage struct {
	db *sqlx.DB
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
	_, err := p.db.Exec(getQuery("save-url"), shortURL, shortURL, originalURL)
	if err != nil {
		return fmt.Errorf("error saving URL: %w", err)
	}
	return nil
}

func init() {
	queryMap = parseQueries(queries)
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

func NewPostgresStorage(dsn string) (*PostgresStorage, error) {
	if dsn == "" {
		return nil, fmt.Errorf("database DSN is empty")
	}

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := applyMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to apply migrations: %w", err)
	}

	return &PostgresStorage{db: db}, nil
}

func (p *PostgresStorage) Ping(ctx context.Context) error {
	return p.db.PingContext(ctx)
}

func createMigrationsTable(db *sqlx.DB) error {
	_, err := db.Exec(getQuery("create-migrations-table"))
	return err
}

func isMigrationApplied(db *sqlx.DB, name string) (bool, error) {
	var count int
	err := db.QueryRow(getQuery("check-migration-applied"), name).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func recordMigration(db *sqlx.DB, name string) error {
	_, err := db.Exec(getQuery("record-migration"), name)
	return err
}

func extractUpSQL(content []byte) (string, error) {
	sqlContent := string(content)

	upParts := strings.Split(sqlContent, "-- +goose Down")
	if len(upParts) > 0 {
		upSQL := strings.Replace(upParts[0], "-- +goose Up", "", 1)
		return strings.TrimSpace(upSQL), nil
	}

	return "", fmt.Errorf("invalid migration format")
}

func applyMigrations(db *sqlx.DB) error {
	if err := createMigrationsTable(db); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	migrationsDir := "migrations"
	_, err := os.Stat(migrationsDir)
	if os.IsNotExist(err) {
		logger.Sugar.Info("Migrations directory does not exist, using built-in schema")
	}

	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		logger.Sugar.Infow("Error reading migrations directory", "error", err)
	}

	if len(files) == 0 {
		logger.Sugar.Info("No migration files found, using built-in schema")
	}

	var migrationFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
			migrationFiles = append(migrationFiles, file.Name())
		}
	}

	sort.Strings(migrationFiles)

	for _, fileName := range migrationFiles {
		migrationPath := filepath.Join(migrationsDir, fileName)

		applied, err := isMigrationApplied(db, fileName)
		if err != nil {
			return fmt.Errorf("failed to check migration status: %w", err)
		}

		if applied {
			logger.Sugar.Infow("Migration already applied, skipping", "file", fileName)
			continue
		}

		content, err := os.ReadFile(migrationPath)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", fileName, err)
		}

		upSQL, err := extractUpSQL(content)
		if err != nil {
			return fmt.Errorf("failed to extract Up SQL from %s: %w", fileName, err)
		}

		tx, err := db.Beginx()
		if err != nil {
			return fmt.Errorf("failed to start transaction for migration %s: %w", fileName, err)
		}

		_, err = tx.Exec(upSQL)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to apply migration %s: %w", fileName, err)
		}

		_, err = tx.Exec(getQuery("record-migration"), fileName)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %s: %w", fileName, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", fileName, err)
		}

		logger.Sugar.Infow("Applied migration", "file", fileName)
	}

	return nil
}

func (p *PostgresStorage) Close() error {
	return p.db.Close()
}
