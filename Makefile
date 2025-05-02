build:
	GOOS=linux GOARCH=amd64 go build cmd/shortener/main.go

run:
	go run cmd/shortener/main.go

PG_HOST ?= localhost
PG_PORT ?= 6432
PG_DB ?= postgres
PG_USER ?= postgres
PG_PASS ?= root
PG_SSL ?= disable

run-postgres:
	DATABASE_DSN="postgres://$(PG_USER):$(PG_PASS)@$(PG_HOST):$(PG_PORT)/$(PG_DB)?sslmode=$(PG_SSL)" go run cmd/shortener/main.go

test:
	go test ./... -v