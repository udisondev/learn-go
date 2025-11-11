.PHONY: help dev run-web run-executor run-verificator build-web build-executor build-verificator generate test test-unit test-integration clean docker-up docker-down db-migrate-up db-migrate-down db-migrate-create db-reset tailwind-watch tailwind-build

# Default target
help:
	@echo "Available commands:"
	@echo "  make dev              - Run with hot reload (air + tailwind watch)"
	@echo "  make run-web          - Run web application"
	@echo "  make run-executor     - Run executor service"
	@echo "  make run-verificator  - Run email verificator service"
	@echo "  make build-web        - Build web application"
	@echo "  make build-executor   - Build executor service"
	@echo "  make build-verificator - Build email verificator service"
	@echo "  make generate         - Generate code (enums, etc)"
	@echo "  make tailwind-build   - Build Tailwind CSS"
	@echo "  make tailwind-watch   - Watch and build Tailwind CSS"
	@echo "  make test             - Run all tests"
	@echo "  make test-unit        - Run unit tests"
	@echo "  make test-integration - Run integration tests"
	@echo "  make docker-up        - Start Docker containers"
	@echo "  make docker-down      - Stop Docker containers"
	@echo "  make db-migrate-up    - Run database migrations"
	@echo "  make db-migrate-down  - Rollback database migrations"
	@echo "  make db-migrate-create NAME=migration_name - Create new migration"
	@echo "  make db-reset         - Reset database (down + up)"
	@echo "  make clean            - Clean build artifacts"

# Run
run-web:
	@echo "Starting web application..."
	go run cmd/web/main.go

run-executor:
	@echo "Starting executor service..."
	go run cmd/executor/main.go

run-verificator:
	@echo "Starting email verificator service..."
	go run cmd/verificator/main.go

# Build
build-web:
	@echo "Building web application..."
	go build -o bin/web cmd/web/main.go

build-executor:
	@echo "Building executor service..."
	go build -o bin/executor cmd/executor/main.go

build-verificator:
	@echo "Building email verificator service..."
	go build -o bin/verificator cmd/verificator/main.go

# Generate
generate:
	@echo "Generating code..."
	go generate ./...

# Development
dev:
	@echo "Starting development mode with hot reload..."
	@make -j2 air tailwind-watch

air:
	@air

tailwind-watch:
	@tailwindcss -i web/static/css/input.css -o web/static/css/output.css --watch

tailwind-build:
	@echo "Building Tailwind CSS..."
	@tailwindcss -i web/static/css/input.css -o web/static/css/output.css --minify

# Test
test:
	@echo "Running all tests..."
	go test -v ./...

test-unit:
	@echo "Running unit tests..."
	go test -v ./tests/unit/...

test-integration:
	@echo "Running integration tests..."
	go test -v ./tests/integration/...

# Docker
docker-up:
	@echo "Starting Docker containers..."
	docker-compose up -d

docker-down:
	@echo "Stopping Docker containers..."
	docker-compose down

# Database migrations (using goose)
db-migrate-up:
	@echo "Running migrations..."
	goose -dir migrations postgres "host=localhost port=5432 user=postgres password=postgres dbname=learn_go sslmode=disable" up

db-migrate-down:
	@echo "Rolling back migrations..."
	goose -dir migrations postgres "host=localhost port=5432 user=postgres password=postgres dbname=learn_go sslmode=disable" down

db-migrate-create:
	@echo "Creating migration: $(NAME)"
	goose -dir migrations -s create $(NAME) sql

db-reset:
	@echo "Resetting database..."
	goose -dir migrations postgres "host=localhost port=5432 user=postgres password=postgres dbname=learn_go sslmode=disable" reset
	goose -dir migrations postgres "host=localhost port=5432 user=postgres password=postgres dbname=learn_go sslmode=disable" up

# Clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -f *.log
