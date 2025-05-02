build:
	GOOS=linux GOARCH=amd64 go build cmd/shortener/main.go

run:
	go run cmd/shortener/main.go

test:
	go test ./... -v