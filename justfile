set dotenv-load := true

default:
    @just --list --unsorted

run:
    go run .

build:
    go build -o server .

test:
    go test -v ./...

tidy:
    go mod tidy

db-up:
    docker compose up -d postgres

db-down:
    docker compose down

db-status:
    docker compose ps

db-logs:
    docker compose logs -f postgres

db-shell:
    docker compose exec postgres psql -U nebula -d nebula_platform

migration name:
    GOOSE_DRIVER=postgres GOOSE_DBSTRING="$DATABASE_URL" GOOSE_MIGRATION_DIR=migrations goose create "{{name}}" sql

migrate-up:
    GOOSE_DRIVER=postgres GOOSE_DBSTRING="$DATABASE_URL" GOOSE_MIGRATION_DIR=migrations goose up

migrate-down:
    GOOSE_DRIVER=postgres GOOSE_DBSTRING="$DATABASE_URL" GOOSE_MIGRATION_DIR=migrations goose down

migrate-status:
    GOOSE_DRIVER=postgres GOOSE_DBSTRING="$DATABASE_URL" GOOSE_MIGRATION_DIR=migrations goose status