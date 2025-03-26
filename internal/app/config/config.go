package config

import "flag"

type ServerOption struct {
	RunAddr      string
	ShortURLAddr string
}

func NewServerOption() *ServerOption {
	opts := &ServerOption{}
	flag.StringVar(&opts.RunAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&opts.ShortURLAddr, "b", "http://localhost:8080", "base address of the resulting shortened URL")

	flag.Parse()

	return opts
}
