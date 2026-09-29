# Basic justfile
default:
    @just --list --unsorted

run:
    go run server.go

build:
    go build -o server server.go

image:
    docker build -t nebula-platform-backend:latest .

container: image
    docker run --rm --name nebula-platform-backend -p 8081:8081 --env-file .env nebula-platform-backend:latest

test:
    go test -v ./...

tidy:
    go mod tidy

