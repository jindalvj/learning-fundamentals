.PHONY: help build run test clean docker-up docker-down docker-logs install

# Default target
help:
	@echo "TinyURL Service (Go) - Available Commands"
	@echo "========================================="
	@echo "  make install      - Download Go dependencies"
	@echo "  make build        - Build the application"
	@echo "  make run          - Run the application locally"
	@echo "  make test         - Run tests"
	@echo "  make lint         - Run golangci-lint"
	@echo "  make fmt          - Format code with gofmt"
	@echo "  make docker-up    - Start Docker services"
	@echo "  make docker-down  - Stop Docker services"
	@echo "  make docker-logs  - View Docker logs"
	@echo "  make clean        - Clean up build artifacts"

# Installation
install:
	go mod download
	go mod verify

# Build
build:
	go build -o bin/tinyurl main.go

# Run application
run:
	go run main.go

# Testing
test:
	go test ./... -v

test-coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out

# Code quality
lint:
	golangci-lint run

fmt:
	gofmt -s -w .
	go mod tidy

# Docker commands
docker-up:
	docker-compose up -d
	@echo "Waiting for services to be ready..."
	@sleep 5
	@docker-compose exec -T postgres pg_isready -U tinyurl_user -d tinyurl
	@echo "Services are ready!"
	@echo "API: http://localhost:8080"

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

docker-rebuild:
	docker-compose down
	docker-compose build --no-cache
	docker-compose up -d

# Cleanup
clean:
	rm -rf bin/
	rm -f coverage.out
	go clean

# Development workflow
dev: docker-up
	@echo "Development environment is ready!"
	@echo "API: http://localhost:8080"
	@echo "Database: localhost:5432"
	@echo "Redis: localhost:6379"

dev-down: docker-down clean
	@echo "Development environment stopped and cleaned up!"
