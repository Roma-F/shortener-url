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
}

type EnvConfig struct {
	ServerAddress   string `env:"SERVER_ADDRESS"`
	ServerPort      string `env:"SERVER_PORT"`
	BaseURL         string `env:"BASE_URL"`
	MaxAttempts     int    `env:"MAX_ATTEMPTS" envDefault:"10"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
}

type flagConfig struct {
	runAddrAlias    string
	runAddr         string
	baseURLAlias    string
	baseURL         string
	fileStoragePath string
}

func parseFlags() flagConfig {
	var fc flagConfig

	flag.StringVar(&fc.runAddrAlias, "a", "", "address and port to run server (alias)")
	flag.StringVar(&fc.runAddr, "server-port", defaultRunAddr, "address and port to run server")

	flag.StringVar(&fc.baseURLAlias, "b", "", "base address for resulting shortened URL (alias)")
	flag.StringVar(&fc.baseURL, "base-url", defaultBaseURL, "base address for resulting shortened URL")

	flag.StringVar(&fc.fileStoragePath, "f", "", "file storage path")
	flag.StringVar(&fc.fileStoragePath, "file-storage", defaultFileStoragePath, "file storage path")

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

	opts := &ServerOption{
		RunAddr:      runAddr,
		ShortURLAddr: baseURL,
		MaxAttempts:  ec.MaxAttempts,
		FSPath:       fsPath,
	}

	return opts, nil
}
