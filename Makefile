.PHONY: dev build run tidy migrate-up migrate-down migrate-version migrate-create

## Install dependencies
tidy:
	go mod tidy

## Run server (applies pending migrations automatically)
dev:
	air

## Build binary
build:
	go build -o bin/warung-ku .

## Run built binary
run:
	./bin/warung-ku

## Install Air for hot reload
install-air:
	go install github.com/air-verse/air@latest

## Install golang-migrate CLI tool
install-migrate:
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

## ── Migration commands ────────────────────────────────────────────────────

## Apply all pending migrations
migrate-up:
	go run main.go -migrate up

## Rollback last migration (use: make migrate-down N=3 to rollback 3)
migrate-down:
	go run main.go -migrate down $(N)

## Show current migration version
migrate-version:
	go run main.go -migrate version

## Create a new migration pair (use: make migrate-create NAME=add_column_x)
## Requires golang-migrate CLI: make install-migrate
migrate-create:
	migrate create -ext sql -dir db/migrations -seq $(NAME)
