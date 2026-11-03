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
    docker compose up -d --wait --wait-timeout 60 postgres

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

# Prepare local dependencies, start Postgres, and apply pending migrations.
setup:
    #!/usr/bin/env bash
    set -euo pipefail

    for tool in go node npm docker goose; do
        if ! command -v "$tool" >/dev/null 2>&1; then
            echo "Missing prerequisite: $tool. See README.md for installation steps." >&2
            exit 1
        fi
    done

    if [[ ! -f .env ]]; then
        cp .env.template .env
        echo "Created .env from .env.template; existing files are never overwritten."
        echo "Fill in the Google credentials supplied by the project maintainer, then rerun just setup."
        exit 0
    fi

    if [[ -z "${DATABASE_URL:-}" ]]; then
        echo "DATABASE_URL is missing. Check .env against .env.template." >&2
        exit 1
    fi

    docker compose version >/dev/null
    if ! docker info >/dev/null 2>&1; then
        echo "Docker is not running or is not accessible. Start Docker and rerun just setup." >&2
        exit 1
    fi

    go mod download
    npm --prefix test-frontend ci

    just --quiet db-up
    just --quiet migrate-up

    echo "Setup complete. Run just run and just frontend in separate terminals."

# Start the test frontend on its expected OAuth redirect port.
frontend:
    npm --prefix test-frontend run dev -- --strictPort

# Regenerate the Swagger specification from controller annotations.
docs:
    go generate .
