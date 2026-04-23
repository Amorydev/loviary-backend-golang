#!/bin/bash
# Development setup script

set -e

echo "=== Loviary Backend Development Setup ==="

# Check if .env exists
if [ ! -f .env ]; then
    echo "Creating .env from .env.example..."
    cp .env.example .env
    echo "Please edit .env with your configuration"
fi

# Install development tools
echo "Installing development tools..."
go install github.com/cosmtrek/air@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Download dependencies
echo "Downloading dependencies..."
go mod download
go mod tidy

# Build the project
echo "Building project..."
make build

echo ""
echo "Setup complete!"
echo ""
echo "Next steps:"
echo "1. Edit .env with your database configuration"
echo "2. Start services: make docker-compose-up"
echo "3. Run migrations: make migrate-up"
echo "4. Access API at http://localhost:8080"
echo ""
