.PHONY: run test fmt vet tidy

run:
	go run ./cmd/api

test:
	go test ./... -short

itest:
	DB_DSN?=host=127.0.0.1 user=postgres password=canandra10 dbname=go_backend_101 port=5432 sslmode=disable TimeZone=Asia/Jakarta go test ./... -run Integration
fmt:
	go fmt ./...

vet:
	go vet ./...

tidy:
	go mod tidy
