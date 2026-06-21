# Load DB settings from .env (ignored if missing, e.g. in CI where env is preset)
-include .env
export

MIGRATIONS_DIR := migrations
DB_URL := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)

deps:
	go mod download
	go mod tidy

docs:
	$(shell go env GOPATH)/bin/swag init -g cmd/api/main.go -o docs

run: docs
	go run ./cmd/api

build: docs
	go build -o bin/api ./cmd/api

clean:
	rm -rf bin/

# Regenerate type-safe query code from migrations + queries/*.sql
sqlc:
	$(shell go env GOPATH)/bin/sqlc generate

# Install the sqlc CLI (only needed once per machine)
sqlc-install:
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# --- Database migrations (golang-migrate) ---

# Install the migrate CLI (only needed once per machine)
migrate-install:
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Create a new migration pair: make migrate-create name=add_orders_table
migrate-create:
	@if [ -z "$(name)" ]; then echo "usage: make migrate-create name=<migration_name>"; exit 1; fi
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)

# Apply all up migrations
migrate-up:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" up

# Roll back the last migration
migrate-down:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" down 1

# Show the current migration version
migrate-version:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" version

# Force-set the version (recover from a dirty state): make migrate-force version=1
migrate-force:
	@if [ -z "$(version)" ]; then echo "usage: make migrate-force version=<n>"; exit 1; fi
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" force $(version)

.PHONY: deps docs run build clean sqlc sqlc-install migrate-install migrate-create migrate-up migrate-down migrate-version migrate-force
