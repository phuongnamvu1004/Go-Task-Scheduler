# Common development tasks. Run `make` or `make help` for the list.
# Recipe lines must be indented with a tab, not spaces.

# Same DB settings as the app: values from .env, then the app's defaults.
-include .env
DB_HOST    ?= localhost
DB_PORT    ?= 5432
DB_SCHEMA  ?= go_task_scheduler
DB_SSLMODE ?= disable

# goose CLI, pinned to the library version in go.mod. It reads its settings
# from these variables, so the password never shows up in the command line.
GOOSE := go run github.com/pressly/goose/v3/cmd/goose@v3.29.0
export GOOSE_DRIVER        := postgres
export GOOSE_MIGRATION_DIR := internal/platform/postgres/migrations
# search_path makes goose use the app's schema and its goose_db_version table.
export GOOSE_DBSTRING = host=$(DB_HOST) port=$(DB_PORT) user=$(DB_USER) password='$(DB_PASSWORD)' dbname=$(DB_NAME) sslmode=$(DB_SSLMODE) search_path=$(DB_SCHEMA)

.DEFAULT_GOAL := help
.PHONY: help run run-debug check swag db-up db-down migrate-create migrate-up migrate-status migrate-down

help: ## List the targets
	@grep -hE '^[a-z-]+:.*## ' $(MAKEFILE_LIST) | awk -F ':.*## ' '{printf "  %-16s %s\n", $$1, $$2}'

run: ## Run the app
	go run ./cmd/scheduler

run-debug: ## Run the app with /debug/scheduler endpoints
	go run -tags debug ./cmd/scheduler

check: ## Build and vet, with and without the debug tag
	go build ./... && go vet ./... && go build -tags debug ./... && go vet -tags debug ./...

swag: ## Regenerate the Swagger docs
	go tool swag init -g cmd/scheduler/main.go --parseInternal -o docs

db-up: ## Start local Postgres
	docker compose up -d

db-down: ## Stop local Postgres (data is kept)
	docker compose down

migrate-create: ## Create the next migration file: make migrate-create name=add_timezone
	@test -n "$(name)" || { echo "usage: make migrate-create name=add_timezone"; exit 1; }
	$(GOOSE) -s create $(name) sql

migrate-up: ## Apply pending migrations (the app also does this on startup)
	$(GOOSE) up

migrate-status: ## Show which migrations are applied
	$(GOOSE) status

migrate-down: ## Undo the latest migration (prints a warning, 5s to cancel)
	@echo "WARNING: rolling back the latest migration on $(DB_NAME) at $(DB_HOST):$(DB_PORT) (schema $(DB_SCHEMA))."
	@echo "Its Down section runs and may drop tables or columns with their data. Ctrl+C within 5s to cancel."
	@sleep 5
	$(GOOSE) down
