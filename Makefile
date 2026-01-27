.PHONY: dev dev-backend dev-frontend install test build clean help wails-build-frontend wails-dev wails-build wails-build-all wails-build-windows wails-build-linux wails-build-darwin wails-build-dmg wails-clean

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

wails-build-frontend: ## Build frontend for Wails (required before wails-build)
	@echo "$(BLUE)🏗️  Building frontend for Wails...$(NC)"
	@cd web && npm run build
	@if [ ! -d "web/dist" ]; then \
		echo "$(YELLOW)⚠️  Warning: web/dist directory not found after build$(NC)"; \
		exit 1; \
	fi
	@echo "$(GREEN)✅ Frontend built successfully$(NC)"
	@echo "$(BLUE)📦 Copying frontend to wails/frontend for embed...$(NC)"
	@rm -rf wails/frontend
	@cp -r web/dist wails/frontend
	@echo "$(GREEN)✅ Frontend copied for Wails embed$(NC)"

# Find Wails CLI - check common locations
WAILS_CMD := $(shell which wails 2>/dev/null || [ -f ~/go/bin/wails ] && echo ~/go/bin/wails || [ -f $(GOPATH)/bin/wails ] && echo $(GOPATH)/bin/wails || echo wails)

# Find Wails CLI - check PATH first, then common Go bin locations
WAILS_CMD := $(shell command -v wails 2>/dev/null || [ -f ~/go/bin/wails ] && echo ~/go/bin/wails || [ -n "$$GOPATH" ] && [ -f $$GOPATH/bin/wails ] && echo $$GOPATH/bin/wails || echo wails)

wails-dev: wails-build-frontend ## Run Wails in development mode with hot reload
	@echo "$(BLUE)🚀 Starting Wails development mode...$(NC)"
	@if ! command -v wails >/dev/null 2>&1 && [ ! -f ~/go/bin/wails ]; then \
		echo "$(YELLOW)⚠️  Wails CLI not found. Please install with:$(NC)"; \
		echo "$(YELLOW)   go install github.com/wailsapp/wails/v2/cmd/wails@latest$(NC)"; \
		echo "$(YELLOW)   Then add ~/go/bin to your PATH$(NC)"; \
		exit 1; \
	fi
	@cd wails && PATH="$$HOME/go/bin:$$PATH" wails dev

wails-build: wails-build-frontend ## Build Wails app for current platform
	@echo "$(BLUE)🏗️  Building Wails app for current platform...$(NC)"
	@if ! command -v wails >/dev/null 2>&1 && [ ! -f ~/go/bin/wails ]; then \
		echo "$(YELLOW)⚠️  Wails CLI not found. Please install with:$(NC)"; \
		echo "$(YELLOW)   go install github.com/wailsapp/wails/v2/cmd/wails@latest$(NC)"; \
		echo "$(YELLOW)   Then add ~/go/bin to your PATH$(NC)"; \
		exit 1; \
	fi
	@cd wails && PATH="$$HOME/go/bin:$$PATH" wails build

wails-build-all: wails-build-frontend ## Build Wails app for all platforms (Windows, Linux, macOS)
	@echo "$(BLUE)🏗️  Building Wails app for all platforms...$(NC)"
	@if ! command -v wails >/dev/null 2>&1 && [ ! -f ~/go/bin/wails ]; then \
		echo "$(YELLOW)⚠️  Wails CLI not found. Please install with:$(NC)"; \
		echo "$(YELLOW)   go install github.com/wailsapp/wails/v2/cmd/wails@latest$(NC)"; \
		echo "$(YELLOW)   Then add ~/go/bin to your PATH$(NC)"; \
		exit 1; \
	fi
	@cd wails && PATH="$$HOME/go/bin:$$PATH" wails build -platform windows/amd64,linux/amd64,darwin/amd64

wails-build-windows: wails-build-frontend ## Build Wails app for Windows
	@echo "$(BLUE)🏗️  Building Wails app for Windows...$(NC)"
	@if ! command -v wails >/dev/null 2>&1 && [ ! -f ~/go/bin/wails ]; then \
		echo "$(YELLOW)⚠️  Wails CLI not found. Please install with:$(NC)"; \
		echo "$(YELLOW)   go install github.com/wailsapp/wails/v2/cmd/wails@latest$(NC)"; \
		echo "$(YELLOW)   Then add ~/go/bin to your PATH$(NC)"; \
		exit 1; \
	fi
	@cd wails && PATH="$$HOME/go/bin:$$PATH" wails build -platform windows/amd64

wails-build-linux: wails-build-frontend ## Build Wails app for Linux
	@echo "$(BLUE)🏗️  Building Wails app for Linux...$(NC)"
	@if ! command -v wails >/dev/null 2>&1 && [ ! -f ~/go/bin/wails ]; then \
		echo "$(YELLOW)⚠️  Wails CLI not found. Please install with:$(NC)"; \
		echo "$(YELLOW)   go install github.com/wailsapp/wails/v2/cmd/wails@latest$(NC)"; \
		echo "$(YELLOW)   Then add ~/go/bin to your PATH$(NC)"; \
		exit 1; \
	fi
	@cd wails && PATH="$$HOME/go/bin:$$PATH" wails build -platform linux/amd64

wails-build-darwin: wails-build-frontend ## Build Wails app for macOS
	@echo "$(BLUE)🏗️  Building Wails app for macOS...$(NC)"
	@if ! command -v wails >/dev/null 2>&1 && [ ! -f ~/go/bin/wails ]; then \
		echo "$(YELLOW)⚠️  Wails CLI not found. Please install with:$(NC)"; \
		echo "$(YELLOW)   go install github.com/wailsapp/wails/v2/cmd/wails@latest$(NC)"; \
		echo "$(YELLOW)   Then add ~/go/bin to your PATH$(NC)"; \
		exit 1; \
	fi
	@cd wails && PATH="$$HOME/go/bin:$$PATH" wails build -platform darwin/amd64

wails-build-dmg: wails-build-darwin ## Create DMG file for macOS distribution
	@echo "$(BLUE)📦 Creating DMG file...$(NC)"
	@APP_NAME="constrictor-rest-client" && \
	APP_PATH="wails/build/bin/darwin/amd64/$$APP_NAME.app" && \
	DMG_NAME="$$APP_NAME.dmg" && \
	TEMP_DMG="$$APP_NAME-temp.dmg" && \
	DMG_DIR="$$APP_NAME-dmg" && \
	if [ ! -d "$$APP_PATH" ]; then \
		echo "$(YELLOW)⚠️  App not found at $$APP_PATH. Building first...$(NC)"; \
		$(MAKE) wails-build-darwin; \
	fi && \
	rm -rf "$$DMG_DIR" "$$DMG_NAME" "$$TEMP_DMG" && \
	mkdir -p "$$DMG_DIR" && \
	cp -R "$$APP_PATH" "$$DMG_DIR/" && \
	ln -s /Applications "$$DMG_DIR/Applications" && \
	hdiutil create -srcfolder "$$DMG_DIR" -volname "$$APP_NAME" -fs HFS+ -fsargs "-c c=64,a=16,e=16" -format UDRW -size 200m "$$TEMP_DMG" && \
	DEVICE=$$(hdiutil attach -readwrite -noverify -noautoopen "$$TEMP_DMG" | egrep '^/dev/' | sed 1q | awk '{print $$1}') && \
	sleep 2 && \
	chmod -Rf go-w "/Volumes/$$APP_NAME" || true && \
	sync && \
	hdiutil detach "$$DEVICE" && \
	hdiutil convert "$$TEMP_DMG" -format UDZO -imagekey zlib-level=9 -o "$$DMG_NAME" && \
	rm -rf "$$DMG_DIR" "$$TEMP_DMG" && \
	echo "$(GREEN)✅ DMG created: $$DMG_NAME$(NC)"

wails-clean: ## Clean Wails build artifacts
	@echo "$(YELLOW)🧹 Cleaning Wails build artifacts...$(NC)"
	@rm -rf wails/build
	@rm -rf web/dist
	@rm -f wails/frontend
	@rm -f *.dmg
	@rm -rf *-dmg
