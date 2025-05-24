package config

import (
	"flag"
	"fmt"
	"strings"

	"github.com/caarlos0/env/v6"
)

const (
	defaultRunAddr         = ":8080"
	defaultBaseURL         = "http://localhost:8080"
	defaultFileStoragePath = "storage.json"
	defaultMigrationsPath  = "migrations"
	defaultMigrationsTable = "migrations"
)

type ServerOption struct {
	RunAddr         string
	ShortURLAddr    string
	MaxAttempts     int
	FSPath          string
	LoggingLevel    string
	DatabaseDSN     string
	MigrationsPath  string
	MigrationsTable string
	ApplyMigrations bool
}

func (o *ServerOption) String() string {
	return fmt.Sprintf(
		"Server Configuration:\n"+
			"  Run Address: %s\n"+
			"  Short URL Address: %s\n"+
			"  File Storage Path: %s\n"+
			"  Max Attempts: %d\n"+
			"  Logging Level: %s\n"+
			"  Database DSN: %s\n"+
			"  Migrations Path: %s\n"+
			"  Migrations Table: %s\n"+
			"  Apply Migrations: %t",
		o.RunAddr,
		o.ShortURLAddr,
		o.FSPath,
		o.MaxAttempts,
		o.LoggingLevel,
		o.DatabaseDSN,
		o.MigrationsPath,
		o.MigrationsTable,
		o.ApplyMigrations,
	)
}

type EnvConfig struct {
	ServerAddress   string `env:"SERVER_ADDRESS"`
	ServerPort      string `env:"SERVER_PORT"`
	BaseURL         string `env:"BASE_URL"`
	MaxAttempts     int    `env:"MAX_ATTEMPTS" envDefault:"10"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	LoggingLevel    string `env:"LOGGING_LEVEL" envDefault:"info"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
	MigrationsPath  string `env:"MIGRATIONS_PATH" envDefault:"migrations"`
	MigrationsTable string `env:"MIGRATIONS_TABLE" envDefault:"migrations"`
	ApplyMigrations bool   `env:"APPLY_MIGRATIONS" envDefault:"true"`
}

type flagConfig struct {
	runAddrAlias    string
	runAddr         string
	baseURLAlias    string
	baseURL         string
	fileStoragePath string
	databaseDSN     string
	migrationsPath  string
	migrationsTable string
	applyMigrations bool
}

func parseFlags() flagConfig {
	var fc flagConfig

	flag.StringVar(&fc.runAddrAlias, "a", "", "address and port to run server (alias)")
	flag.StringVar(&fc.runAddr, "server-port", defaultRunAddr, "address and port to run server")

	flag.StringVar(&fc.baseURLAlias, "b", "", "base address for resulting shortened URL (alias)")
	flag.StringVar(&fc.baseURL, "base-url", defaultBaseURL, "base address for resulting shortened URL")

	flag.StringVar(&fc.fileStoragePath, "f", "", "file storage path")
	flag.StringVar(&fc.fileStoragePath, "file-storage", defaultFileStoragePath, "file storage path")

	flag.StringVar(&fc.databaseDSN, "d", "", "database connection string")

	flag.StringVar(&fc.migrationsPath, "migrations-path", defaultMigrationsPath, "path to migrations directory")
	flag.StringVar(&fc.migrationsTable, "migrations-table", defaultMigrationsTable, "name of migrations table")
	flag.BoolVar(&fc.applyMigrations, "apply-migrations", true, "automatically apply migrations on startup")

	flag.Parse()
	return fc
}

func parseEnv() (EnvConfig, error) {
	var ec EnvConfig
	if err := env.Parse(&ec); err != nil {
		return EnvConfig{}, fmt.Errorf("failed to parse environment variables: %w", err)
	}
	return ec, nil
}

func NewServerOption() (*ServerOption, error) {
	fc := parseFlags()
	ec, err := parseEnv()
	if err != nil {
		return nil, err
	}

	runAddr := fc.runAddr
	if fc.runAddrAlias != "" {
		runAddr = fc.runAddrAlias
	}
	if ec.ServerAddress != "" {
		runAddr = ec.ServerAddress
	} else if ec.ServerPort != "" {
		if !strings.HasPrefix(ec.ServerPort, ":") {
			runAddr = ":" + ec.ServerPort
		} else {
			runAddr = ec.ServerPort
		}
	}

	fsPath := defaultFileStoragePath
	if fc.fileStoragePath != defaultFileStoragePath {
		fsPath = fc.fileStoragePath
	}
	if ec.FileStoragePath != "" {
		fsPath = ec.FileStoragePath
	}

	baseURL := fc.baseURL
	if fc.baseURLAlias != "" {
		baseURL = fc.baseURLAlias
	}
	if ec.BaseURL != "" {
		baseURL = ec.BaseURL
	}

	loggingLevel := ec.LoggingLevel

	databaseDSN := fc.databaseDSN
	if ec.DatabaseDSN != "" {
		databaseDSN = ec.DatabaseDSN
	}

	migrationsPath := fc.migrationsPath
	if ec.MigrationsPath != "" {
		migrationsPath = ec.MigrationsPath
	}

	migrationsTable := fc.migrationsTable
	if ec.MigrationsTable != "" {
		migrationsTable = ec.MigrationsTable
	}

	applyMigrations := ec.ApplyMigrations
	if !fc.applyMigrations {
		applyMigrations = false
	}

	opts := &ServerOption{
		RunAddr:         runAddr,
		ShortURLAddr:    baseURL,
		MaxAttempts:     ec.MaxAttempts,
		FSPath:          fsPath,
		LoggingLevel:    loggingLevel,
		DatabaseDSN:     databaseDSN,
		MigrationsPath:  migrationsPath,
		MigrationsTable: migrationsTable,
		ApplyMigrations: applyMigrations,
	}

	return opts, nil
}
