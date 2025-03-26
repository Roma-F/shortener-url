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

	var runAddrAlias string
	var runAddrFlag string
	flag.StringVar(&runAddrAlias, "a", "", "address and port to run server (alias)")
	flag.StringVar(&runAddrFlag, "server-port", defaultRunAddr, "address and port to run server")

	var baseURLAlias string
	var baseURLFlag string
	flag.StringVar(&baseURLAlias, "b", "", "base address for resulting shortened URL (alias)")
	flag.StringVar(&baseURLFlag, "base-url", defaultBaseURL, "base address for resulting shortened URL")

	flag.Parse()

	finalRunAddr := runAddrFlag
	if runAddrAlias != "" {
		finalRunAddr = runAddrAlias
	}
	finalBaseURL := baseURLFlag
	if baseURLAlias != "" {
		finalBaseURL = baseURLAlias
	}

	opts := &ServerOption{
		RunAddr:      finalRunAddr,
		ShortURLAddr: finalBaseURL,
	}

	if v, ok := os.LookupEnv("SERVER_ADDRESS"); ok {
		opts.RunAddr = v
	} else if v, ok := os.LookupEnv("SERVER_PORT"); ok {
		if !strings.HasPrefix(v, ":") {
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
