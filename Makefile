.PHONY: build run test clean migrate seed

# Build the application
build:
	go build -o bin/app cmd/main.go

# Run the application
run:
	go run cmd/main.go

# Run tests
test:
	go test -v ./test/...

# Clean build artifacts
clean:
	rm -rf bin/

# Run database migrations
migrate:
	go run cmd/main.go migrate

# Seed database with initial data
seed:
	go run cmd/main.go seed

# Run development server with hot reload
dev:
	air

# Format code
fmt:
	go fmt ./...

# Run code linter
lint:
	go vet ./...

# Download dependencies
deps:
	go mod download

# Update dependencies
deps-update:
	go get -u ./...

# Show help
help:
	@echo "Available commands:"
	@echo "  build        - Build the application"
	@echo "  run          - Run the application"
	@echo "  test         - Run tests"
	@echo "  clean        - Clean build artifacts"
	@echo "  migrate      - Run database migrations"
	@echo "  seed         - Seed database with initial data"
	@echo "  dev          - Run development server with hot reload"
	@echo "  fmt          - Format code"
	@echo "  lint         - Run code linter"
	@echo "  deps         - Download dependencies"
	@echo "  deps-update  - Update dependencies"
	@echo "  help         - Show this help message"