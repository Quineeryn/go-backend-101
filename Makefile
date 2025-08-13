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


MIGRATE?=migrate
DB_URL?=postgres://postgres:postgres@127.0.0.1:5432/go_backend_101?sslmode=disable

migrate-up:
	$(MIGRATE) -path db/migrations -database "$(DB_URL)" up

migrate-down:
	$(MIGRATE) -path db/migrations -database "$(DB_URL)" down 1
