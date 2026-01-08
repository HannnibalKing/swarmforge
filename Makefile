.PHONY: help install dev build test clean migrate docker-up docker-down

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

install: ## Install all dependencies
	@echo "Installing Go dependencies..."
	cd services/api-gateway && go mod download
	cd services/job-orchestrator && go mod download
	cd services/printer-service && go mod download
	cd services/model-processor && go mod download
	@echo "Installing Node.js dependencies..."
	cd services/qa-service && npm install
	cd services/websocket-server && npm install
	cd frontend && npm install
	@echo "Dependencies installed!"

dev: ## Start development environment
	docker-compose up

docker-up: ## Start all services in Docker
	docker-compose up -d
	@echo "Services started! Access:"
	@echo "  Frontend:  http://localhost:3000"
	@echo "  API:       http://localhost:8080"
	@echo "  MinIO UI:  http://localhost:9001"
	@echo "  RabbitMQ:  http://localhost:15672"

docker-down: ## Stop all services
	docker-compose down

docker-clean: ## Stop services and remove volumes
	docker-compose down -v

migrate: ## Run database migrations
	@echo "Running migrations..."
	docker-compose exec postgres psql -U swarmforge -d swarmforge -f /docker-entrypoint-initdb.d/01_schema.sql
	@echo "Migrations complete!"

test: ## Run all tests
	@echo "Running Go tests..."
	cd services/api-gateway && go test ./...
	cd services/job-orchestrator && go test ./...
	cd services/printer-service && go test ./...
	@echo "Running Node.js tests..."
	cd services/qa-service && npm test
	cd services/websocket-server && npm test
	cd frontend && npm test

build: ## Build all services
	@echo "Building services..."
	docker-compose build

clean: ## Clean build artifacts
	@echo "Cleaning..."
	find . -name "node_modules" -type d -exec rm -rf {} +
	find . -name "dist" -type d -exec rm -rf {} +
	find . -name ".next" -type d -exec rm -rf {} +

logs: ## View logs from all services
	docker-compose logs -f

setup-buckets: ## Create MinIO buckets
	docker-compose exec minio mc alias set local http://localhost:9000 minioadmin minioadmin
	docker-compose exec minio mc mb local/swarmforge-models
	docker-compose exec minio mc mb local/swarmforge-qa-photos
	docker-compose exec minio mc policy set download local/swarmforge-models
	docker-compose exec minio mc policy set download local/swarmforge-qa-photos
