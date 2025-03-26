package config

import (
	"flag"
	"os"
)

type ServerOption struct {
	RunAddr      string
	ShortURLAddr string
}

func NewServerOption() *ServerOption {
	opts := &ServerOption{}
	flag.StringVar(&opts.RunAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&opts.ShortURLAddr, "b", "http://localhost:8080", "base address of the resulting shortened URL")

	if v, ok := os.LookupEnv("SERVER_ADDRESS"); ok {
		opts.RunAddr = v
	}
	if v, ok := os.LookupEnv("BASE_URL"); ok {
		opts.ShortURLAddr = v
	}

	return opts
}
