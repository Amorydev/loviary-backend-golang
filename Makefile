.PHONY: help build run test clean migrate migrate-up migrate-down docker-build docker-run deps tidy lint

# Build variables
BINARY_NAME=loviary
BUILD_DIR=bin
CMD_DIR=cmd/api
MAIN=$(CMD_DIR)/main.go

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

tidy: ## Tidy go.mod and go.sum
	@echo "Tidying dependencies..."
	@go mod tidy

build: ## Build the application
	@echo "Building application..."
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN)
	@echo "Binary built at $(BUILD_DIR)/$(BINARY_NAME)"

run: ## Run the application
	@echo "Running application..."
	@go run $(MAIN)

run-dev: ## Run with air for hot reload (requires air: go install github.com/cosmtrek/air@latest)
	@echo "Running with hot reload..."
	@air

test: ## Run tests
	@echo "Running tests..."
	@go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated at coverage.html"

test-integration: ## Run integration tests
	@echo "Running integration tests..."
	@go test -v -tags=integration ./...

lint: ## Run linter
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	fi

vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html

# Database commands
migrate-up: ## Run migrations up
	@echo "Running migrations up..."
	@go run github.com/pressly/goose/v3/cmd/goose@latest -dir migrations postgres "$(DB_URL)" up

migrate-down: ## Run migrations down
	@echo "Running migrations down..."
	@go run github.com/pressly/goose/v3/cmd/goose@latest -dir migrations postgres "$(DB_URL)" down

migrate-status: ## Show migration status
	@echo "Migration status:"
	@go run github.com/pressly/goose/v3/cmd/goose@latest -dir migrations postgres "$(DB_URL)" status

migrate-create: ## Create a new migration (use NAME=name)
	@if [ -z "$(NAME)" ]; then echo "Usage: make migrate-create NAME=migration_name"; exit 1; fi
	@go run github.com/pressly/goose/v3/cmd/goose@latest -dir migrations create $(NAME)
	@echo "Migration created in migrations/"

# Docker commands
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t loviary:latest .

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	@docker run -p 8080:8080 --env-file .env loviary:latest

docker-compose-up: ## Start all services with docker-compose
	@echo "Starting services..."
	@docker-compose up --build

docker-compose-down: ## Stop all services
	@echo "Stopping services..."
	@docker-compose down

docker-compose-down-clean: ## Stop services and remove volumes
	@echo "Stopping services and cleaning volumes..."
	@docker-compose down -v

# Development utilities
dev-setup: deps ## Setup development environment
	@echo "Development setup complete!"
	@echo "Run 'make docker-compose-up' to start the application"

# Quick check before commit
pre-commit: fmt vet lint test ## Run all checks before commit
	@echo "All checks passed!"

# Install required dev tools
install-tools:
	@echo "Installing development tools..."
	@go install github.com/cosmtrek/air@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/go-playground/validator/v10@latest
	@echo "Tools installed!"
