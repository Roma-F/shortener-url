package config

import (
	"flag"
	"fmt"
	"strings"

	"github.com/caarlos0/env/v6"
)

const (
	defaultRunAddr = ":8080"
	defaultBaseURL = "http://localhost:8080"
)

type ServerOption struct {
	RunAddr      string
	ShortURLAddr string
	MaxAttempts  int
}

type EnvConfig struct {
	ServerAddress string `env:"SERVER_ADDRESS"`
	ServerPort    string `env:"SERVER_PORT"`
	BaseURL       string `env:"BASE_URL"`
	MaxAttempts   int    `env:"MAX_ATTEMPTS" envDefault:"10"`
}

type flagConfig struct {
	runAddrAlias string
	runAddr      string
	baseURLAlias string
	baseURL      string
}

func parseFlags() flagConfig {
	var fc flagConfig

	flag.StringVar(&fc.runAddrAlias, "a", "", "address and port to run server (alias)")
	flag.StringVar(&fc.runAddr, "server-port", defaultRunAddr, "address and port to run server")

	flag.StringVar(&fc.baseURLAlias, "b", "", "base address for resulting shortened URL (alias)")
	flag.StringVar(&fc.baseURL, "base-url", defaultBaseURL, "base address for resulting shortened URL")

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
	}

	return opts, nil
}
