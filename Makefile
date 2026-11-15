.PHONY: dev-api dev-web seed test build docker-up docker-down docker-logs

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
	go build ./cmd/server ./cmd/seed
	npm --prefix web run build

docker-up:
	docker compose up --build -d

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f --tail=100
