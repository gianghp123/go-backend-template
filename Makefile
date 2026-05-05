-include .env
export

.PHONY: migrate-create migrate-up migrate-down up down build logs restart clean dev test test-coverage

DOCKER_COMPOSE = docker compose --env-file .env -f docker/docker-compose.yaml

DB_URL = postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)

migrate-create:
	@if [ -z "$(name)" ]; then \
		echo "Usage: make migrate-create name=your_migration_name"; \
		exit 1; \
	fi
	goose -dir internal/database/migrations create $(name) sql

migrate-up:
	goose -dir internal/database/migrations postgres "$(DB_URL)" up

migrate-down:
	goose -dir internal/database/migrations postgres "$(DB_URL)" down

up:
	$(DOCKER_COMPOSE) up -d --build

down:
	$(DOCKER_COMPOSE) down

build:
	$(DOCKER_COMPOSE) build

logs:
	$(DOCKER_COMPOSE) logs -f

restart:
	$(DOCKER_COMPOSE) down && $(DOCKER_COMPOSE) up -d --build

clean:
	$(DOCKER_COMPOSE) down -v

dev:
	# Replace with your hot-reload tool: air, realize, nodemon, etc.
	go run ./cmd/server

test:
	go test ./... -v

test-coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
