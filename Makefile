.PHONY: dev dev-backend dev-frontend install test build clean help

# Default target
.DEFAULT_GOAL := help

# Colors
GREEN := \033[0;32m
BLUE := \033[0;34m
YELLOW := \033[1;33m
NC := \033[0m # No Color

help: ## Show this help message
	@echo "$(BLUE)Constrictor REST Client - Available commands:$(NC)"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-15s$(NC) %s\n", $$1, $$2}'

dev: ## Run both backend and frontend (requires npm install in root)
	@echo "$(BLUE)🚀 Starting Constrictor REST Client...$(NC)"
	@npm run dev

dev-backend: ## Run only the backend server
	@echo "$(GREEN)🔧 Starting backend server on http://localhost:8080...$(NC)"
	@go run cmd/constrictor-rest-client/main.go

dev-frontend: ## Run only the frontend server
	@echo "$(GREEN)🎨 Starting frontend server on http://localhost:5173...$(NC)"
	@cd web && npm run dev

install: ## Install all dependencies (Go and npm)
	@echo "$(YELLOW)📦 Installing dependencies...$(NC)"
	@go mod tidy
	@cd web && npm install
	@npm install

test: ## Run all Go tests
	@echo "$(BLUE)🧪 Running tests...$(NC)"
	@go test ./...

build: ## Build frontend for production
	@echo "$(BLUE)🏗️  Building frontend...$(NC)"
	@cd web && npm run build

clean: ## Clean build artifacts
	@echo "$(YELLOW)🧹 Cleaning...$(NC)"
	@rm -rf web/dist
	@rm -rf web/node_modules
	@rm -rf node_modules
