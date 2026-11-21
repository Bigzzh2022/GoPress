.PHONY: dev-api dev-web seed test build docker-up docker-down docker-logs

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT ?= $(shell git rev-parse HEAD 2>/dev/null || echo unknown)
BUILD_TIME ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -X gopress/internal/buildinfo.Version=$(VERSION) -X gopress/internal/buildinfo.Commit=$(COMMIT) -X gopress/internal/buildinfo.BuildTime=$(BUILD_TIME)

dev-api:
	go run ./cmd/server

dev-web:
	npm --prefix web run dev

seed:
	go run ./cmd/seed

test:
	go test ./...
	npm --prefix web test -- --run

build:
	go build -trimpath -ldflags="$(LDFLAGS)" ./cmd/server ./cmd/seed
	npm --prefix web run build

docker-up:
	docker compose up --build -d

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f --tail=100
