.PHONY: help install run test clean docker-up docker-down docker-logs db-init db-reset

# Default target
help:
	@echo "TinyURL Service - Available Commands"
	@echo "===================================="
	@echo "  make install      - Install Python dependencies"
	@echo "  make run          - Run the application locally"
	@echo "  make test         - Run API tests"
	@echo "  make test-unit    - Run unit tests"
	@echo "  make lint         - Run code linting"
	@echo "  make format       - Format code with black"
	@echo "  make docker-up    - Start Docker services"
	@echo "  make docker-down  - Stop Docker services"
	@echo "  make docker-logs  - View Docker logs"
	@echo "  make db-init      - Initialize database"
	@echo "  make db-reset     - Reset database"
	@echo "  make clean        - Clean up generated files"

# Installation
install:
	pip install -r requirements.txt

# Run application
run:
	python api.py

# Testing
test:
	python test_api.py

test-unit:
	python base62_encoder.py

# Code quality
lint:
	pylint *.py

format:
	black *.py

# Docker commands
docker-up:
	docker-compose up -d
	@echo "Waiting for services to be ready..."
	@sleep 5
	@docker-compose exec -T postgres pg_isready -U tinyurl_user -d tinyurl
	@echo "Services are ready!"

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

docker-rebuild:
	docker-compose down
	docker-compose build --no-cache
	docker-compose up -d

# Database commands
db-init:
	@echo "Initializing database..."
	@if [ -f "schema.sql" ]; then \
		docker-compose exec -T postgres psql -U tinyurl_user -d tinyurl < schema.sql; \
		echo "Database initialized successfully!"; \
	else \
		echo "Error: schema.sql not found"; \
	fi

db-reset:
	@echo "Resetting database..."
	@docker-compose exec -T postgres psql -U tinyurl_user -d tinyurl -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
	@$(MAKE) db-init

db-shell:
	docker-compose exec postgres psql -U tinyurl_user -d tinyurl

# Cleanup
clean:
	find . -type d -name "__pycache__" -exec rm -rf {} +
	find . -type f -name "*.pyc" -delete
	find . -type f -name "*.pyo" -delete
	find . -type f -name "*.log" -delete
	find . -type d -name "*.egg-info" -exec rm -rf {} +
	rm -rf .pytest_cache
	rm -rf htmlcov
	rm -rf .coverage

# Development workflow
dev: docker-up
	@echo "Development environment is ready!"
	@echo "API: http://localhost:5000"
	@echo "Database: localhost:5432"
	@echo "Redis: localhost:6379"

dev-down: docker-down clean
	@echo "Development environment stopped and cleaned up!"
