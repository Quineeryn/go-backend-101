DB_DSN=host=127.0.0.1 user=postgres password=canandra10 dbname=go_backend_101 port=5432 sslmode=disable TimeZone=Asia/Jakarta

.PHONY: run test fmt vet tidy

run:
	CGO_ENABLED=0 go run ./cmd/api

test:
	CGO_ENABLED=0 go test ./internal/... -v

build:
	CGO_ENABLED=0 go build -o bin/api ./cmd/api

itest:
	DB_DSN="$(DB_DSN)" go test -tags=integration ./... -run Integration -v
fmt:
	go fmt ./...

vet:
	go vet ./...

tidy:
	go mod tidy

migrate-pg:
	CGO_ENABLED=0 go run ./cmd/migrate


MIGRATE?=migrate
DB_URL?=postgres://postgres:canandra10@127.0.0.1:5432/go_backend_101?sslmode=disable

migrate-up:
	$(MIGRATE) -path db/migrations -database "$(DB_URL)" up

migrate-down:
	$(MIGRATE) -path db/migrations -database "$(DB_URL)" down 1
