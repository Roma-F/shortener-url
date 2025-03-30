package config

import (
	"flag"
	"log"
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
}

type EnvConfig struct {
	ServerAddress string `env:"SERVER_ADDRESS"`
	ServerPort    string `env:"SERVER_PORT"`
	BaseURL       string `env:"BASE_URL"`
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

func parseEnv() EnvConfig {
	var ec EnvConfig
	if err := env.Parse(&ec); err != nil {
		log.Fatal("failed to parse environment variables: ", err)
	}
	return ec
}

func NewServerOption() *ServerOption {
	fc := parseFlags()
	ec := parseEnv()

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
	}

	log.Printf("Server will run on %s", opts.RunAddr)
	log.Printf("Base URL is %s", opts.ShortURLAddr)
	return opts
}
