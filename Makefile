# Makefile para URL Shortener

# Variáveis
APP_NAME=url-shortener
DOCKER_COMPOSE=docker-compose
GO_CMD=go
BINARY_NAME=url-shortener

# Cores para output
GREEN=\033[0;32m
YELLOW=\033[1;33m
RED=\033[0;31m
NC=\033[0m # No Color

.PHONY: help up up-tools down ps logs postgres-log redis-logs clean restart deps \
	build run dev test test-coverage db-migrate db-reset clean-files fmt lint

# Help
help: ## Show this help
	@echo "$(GREEN)Available commands:$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(YELLOW)%-20s$(NC) %s\n", $$1, $$2}'

# Docker Commands
up: ## Run the containers (postgres and redis)
	@echo "$(GREEN)Running the containers...$(NC)"
	$(DOCKER_COMPOSE) up -d postgres redis

up-tools: ## Run the containers including tools (redis-commander)
	@echo "$(GREEN)Running the containers with tools...$(NC)"
	$(DOCKER_COMPOSE) --profile tools up -d

down: ## Para e remove os containers
	@echo "$(YELLOW)Stopping containers...$(NC)"
	$(DOCKER_COMPOSE) down

ps: ## Para e remove os containers
	$(DOCKER_COMPOSE) ps

logs: ## Mostra os logs dos containers
	$(DOCKER_COMPOSE) logs -f

postgres-logs: ## Show the logs for PostgreSQL
	$(DOCKER_COMPOSE) logs -f postgres

redis-logs: ## Show the logs for Redis
	$(DOCKER_COMPOSE) logs -f redis

clean: ## Remove containers, volumes and images
	@echo "$(RED)Removing containers, volumes and images...$(NC)"
	$(DOCKER_COMPOSE) down -v --remove-orphans
	docker system prune -f

restart: ## Reboot the containers
	@echo "$(YELLOW)Rebooting containers...$(NC)"
	$(DOCKER_COMPOSE) restart

# Go Commands
deps: ## Install dependencies of Go
	@echo "$(GREEN)Installing dependencies...$(NC)"
	$(GO_CMD) mod tidy
	$(GO_CMD) mod download

build: ## Compile the application
	@echo "$(GREEN)Compiling the application...$(NC)"
	$(GO_CMD) build -o bin/$(BINARY_NAME) cmd/server/main.go

run: ## Run the application locally
	@echo "$(GREEN)Running application...$(NC)"
	$(GO_CMD) run cmd/main.go

dev: docker-up deps ## Prepare the developer environment
	@echo "$(GREEN)Developer environment ready!$(NC)"
	@echo "$(YELLOW)PostgreSQL:$(NC) localhost:5432"
	@echo "$(YELLOW)Redis:$(NC) localhost:6379"
	@echo "$(YELLOW)To run the application:$(NC) make run"

test: ## Run the tests
	@echo "$(GREEN)Running tests...$(NC)"
	$(GO_CMD) test -v ./...

test-coverage: ## Run tests with coverage
	@echo "$(GREEN)Running tests with coverage...$(NC)"
	$(GO_CMD) test -v -coverprofile=coverage.out ./...
	$(GO_CMD) tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report generated in coverage.html$(NC)"

# Database Commands
db-migrate: ## Run the migrations ()
	@echo "$(GREEN)Running migrations...$(NC)"
	$(GO_CMD) run cmd/migrate/main.go

db-reset: docker-down ## Reseta o banco de dados
	@echo "$(YELLOW)Reseting the database...$(NC)"
	docker volume rm url-shortener_postgres_data || true
	$(MAKE) docker-up

# Utility Commands
clean-files: ## Clean temporary files and binaries
	@echo "$(YELLOW)Cleaning temporary files...$(NC)"
	rm -rf bin/
	rm -rf tmp/
	rm -f coverage.out coverage.html
	$(GO_CMD) clean

fmt: ## Format the code in Go
	@echo "$(GREEN)Formatting code...$(NC)"
	$(GO_CMD) fmt ./...

lint: ## Run the linter (golangci-lint required)
	@echo "$(GREEN)Running linter...$(NC)"
	golangci-lint run
