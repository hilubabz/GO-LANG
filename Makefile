BINARY_NAME=go-lang

.PHONY: all build run test fmt vet tidy \
	migrate-up migrate-down migrate-create \
	sqlc-generate clean

all: build

# --------------------
# Go
# --------------------

build:
	go build -o bin/$(BINARY_NAME) .

run:
	go run .

test:
	go test ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

tidy:
	go mod tidy

# --------------------
# Database Migrations
# --------------------

migrate-up:
	goose -dir sql/schema postgres "$(DB_URL)" up

migrate-down:
	goose -dir sql/schema postgres "$(DB_URL)" down

migrate-create:
	goose -dir sql/schema postgres "$(DB_URL)" create $(name) sql

# --------------------
# SQLC
# --------------------

sqlc-generate:
	sqlc generate

# --------------------
# Clean
# --------------------

clean:
	rm -rf bin/
