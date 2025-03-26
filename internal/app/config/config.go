package config

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

type ServerOption struct {
	RunAddr      string
	ShortURLAddr string
}

func NewServerOption() *ServerOption {
	defaultRunAddr := ":8080"
	defaultBaseURL := "http://localhost:8080"

	flag.StringVar(&defaultRunAddr, "server-port", defaultRunAddr, "address and port to run server")
	flag.StringVar(&defaultBaseURL, "base-url", defaultBaseURL, "base address for resulting shortened URL")
	flag.Parse()

	opts := &ServerOption{
		RunAddr:      defaultRunAddr,
		ShortURLAddr: defaultBaseURL,
	}

	if v, ok := os.LookupEnv("SERVER_PORT"); ok {
		if !strings.Contains(v, ":") {
			opts.RunAddr = ":" + v
		} else {
			opts.RunAddr = v
		}
	}
	if v, ok := os.LookupEnv("BASE_URL"); ok {
		opts.ShortURLAddr = v
	}

	fmt.Println("Server will run on", opts.RunAddr)
	fmt.Println("Base URL is", opts.ShortURLAddr)
	return opts
}
