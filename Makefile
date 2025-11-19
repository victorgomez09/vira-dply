# Default target
help:
	@echo "Available targets:"
	@echo "  build-server - Build the mikrocloud-server binary with embedded frontend"
	@echo "  build-cli    - Build the mikrocloud-cli binary (management tool)"
	@echo "  build-all    - Build both server and CLI binaries"
	@echo "  build        - Alias for build-all"
	@echo "  build-web    - Build the frontend assets"
	@echo "  run          - Run the mikrocloud server"
	@echo "  clean        - Clean build artifacts"
	@echo "  test         - Run tests"
	@echo "  deps         - Download dependencies"
	@echo "  migrate      - Run database migrations"
	@echo "  migrate-up   - Apply all pending migrations"
	@echo "  migrate-down - Rollback the last migration"
	@echo "  migrate-status - Show migration status"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run with Docker Compose"

# Build the CLI binary (no embedded frontend)
build-cli: deps
	@echo "Building CLI binary..."
	go build -o bin/mikrocloud-cli ./cmd/cli/main.go
	@echo "✅ CLI built successfully at bin/mikrocloud-cli"

# Build the server binary (with embedded frontend)
build-server: build-web deps
	@echo "Building server binary with embedded frontend..."
	go build -o bin/mikrocloud-server ./cmd/api/main.go
	@echo "✅ Server built successfully at bin/mikrocloud-server"

# Build both binaries
build-all: build-server build-cli
	@echo "✅ All binaries built successfully"

# Alias for backward compatibility
build: build-all

# Build the frontend assets
build-web:
	@echo "Building frontend assets..."
	cd web && pnpm install
	cd web && pnpm run build
	@echo "✅ Frontend built successfully at web/dist/"

# Run the server (builds server if binary doesn't exist)
run: bin/mikrocloud-server
	./bin/mikrocloud-server serve

# Ensure the server binary exists
bin/mikrocloud-server:
	$(MAKE) build-server

# Run in development mode with auto-reload (requires air)
dev:
	air

# Clean build artifacts
clean:
	rm -rf bin/
	rm -rf web/dist/
	rm -rf web/.svelte-kit/
	rm -rf web/node_modules/
	go clean

# Run tests
test:
	go test -v ./...

# Download dependencies
deps:
	go mod download
	go mod tidy

# Download frontend dependencies
deps-web:
	@echo "Installing frontend dependencies..."
	cd web && pnpm install

# Download all dependencies
deps-all: deps deps-web

# Database migrations using goose
migrate: migrate-up

migrate-up:
	@mkdir -p $(HOME)/.local/share/mikrocloud
	goose -dir migrations sqlite3 "$(DATABASE_URL)" up

migrate-down:
	goose -dir migrations sqlite3 "$(DATABASE_URL)" down

migrate-status:
	goose -dir migrations sqlite3 "$(DATABASE_URL)" status

# Create a new migration
migrate-create:
	@read -p "Enter migration name: " name; \
	goose -dir migrations create $$name sql

# Docker targets
docker-build:
	docker build -t mikrocloud:latest .

docker-run:
	docker-compose up -d

docker-stop:
	docker-compose down

# Build helper image for builds
build-helper:
	docker build -t mikrocloud-builder:latest ./docker/Build-Helper
	docker tag mikrocloud-builder:latest ghcr.io/fantasy-programming/mikrocloud-2/mikrocloud-builder:latest

build-helper-push:
	docker push ghcr.io/fantasy-programming/mikrocloud-2/mikrocloud-builder:latest

# Development database - SQLite (no external dependencies needed)
db-init:
	@mkdir -p $(HOME)/.local/share/mikrocloud
	@echo "SQLite database will be created automatically"

db-clean:
	rm -f $(HOME)/.local/share/mikrocloud/mikrocloud.db*

# Install tools
install-tools:
	go install github.com/pressly/goose/v3/cmd/goose@latest
	go install github.com/air-verse/air@latest
	npm install -g pnpm

# Development mode - frontend dev server + backend
dev-full:
	@echo "Starting development mode..."
	@echo "Frontend will be available at http://localhost:5173"
	@echo "Backend will be available at http://localhost:3000"
	cd web && pnpm run dev &
	$(MAKE) dev

# Development mode with frontend built and embedded
dev-embedded: build-full
	air

# Format code
fmt:
	go fmt ./...

# Lint code (requires golangci-lint)
lint:
	golangci-lint run

# Run security checks (requires gosec)
security:
	gosec ./...

# Generate code coverage
coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Default environment variables
export DATABASE_URL ?= $(HOME)/.local/share/mikrocloud/mikrocloud.db
export PORT ?= 3000
export LOG_LEVEL ?= info
