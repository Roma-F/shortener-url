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
)

type ServerOption struct {
	RunAddr      string
	ShortURLAddr string
	MaxAttempts  int
	FSPath       string
	LoggingLevel string
	DatabaseDSN  string
}

func (o *ServerOption) String() string {
	return fmt.Sprintf(
		"Server Configuration:\n"+
			"  Run Address: %s\n"+
			"  Short URL Address: %s\n"+
			"  File Storage Path: %s\n"+
			"  Max Attempts: %d\n"+
			"  Logging Level: %s\n"+
			"  Database DSN: %s",
		o.RunAddr,
		o.ShortURLAddr,
		o.FSPath,
		o.MaxAttempts,
		o.LoggingLevel,
		o.DatabaseDSN,
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
}

type DBParams struct {
	Host     string
	User     string
	Password string
	Dbname   string
	SSLmode  bool
}

type flagConfig struct {
	runAddrAlias    string
	runAddr         string
	baseURLAlias    string
	baseURL         string
	fileStoragePath string
	databaseDSN     string
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

	opts := &ServerOption{
		RunAddr:      runAddr,
		ShortURLAddr: baseURL,
		MaxAttempts:  ec.MaxAttempts,
		FSPath:       fsPath,
		LoggingLevel: loggingLevel,
		DatabaseDSN:  fc.databaseDSN,
	}

	if ec.DatabaseDSN != "" {
		opts.DatabaseDSN = ec.DatabaseDSN
	}

	return opts, nil
}
