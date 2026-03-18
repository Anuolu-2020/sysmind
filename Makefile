# SysMind Makefile
# Common development tasks

.PHONY: help install dev build test clean lint release version-check

# Default target
help: ## Show this help message
	@echo "SysMind Development Commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

install: ## Install dependencies
	@echo "📦 Installing dependencies..."
	cd frontend && npm install
	go mod tidy
	@echo "✅ Dependencies installed"

dev: ## Start development server
	@echo "🚀 Starting development server..."
	wails dev

build: ## Build for production
	@echo "🏗️ Building for production..."
	./scripts/build.sh --prod

build-dev: ## Build for development
	@echo "🏗️ Building for development..."
	./scripts/build.sh

test: ## Run all tests
	@echo "🧪 Running tests..."
	go test -v -race ./...
	cd frontend && npm test -- --watchAll=false

test-coverage: ## Run tests with coverage
	@echo "📊 Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

lint: ## Run linters
	@echo "🔍 Running linters..."
	golangci-lint run
	cd frontend && npm run lint || true

lint-fix: ## Run linters and fix issues
	@echo "🔧 Running linters and fixing issues..."
	golangci-lint run --fix
	cd frontend && npm run lint:fix || true

clean: ## Clean build artifacts
	@echo "🧹 Cleaning build artifacts..."
	rm -rf build/
	rm -f coverage.out coverage.html
	cd frontend && rm -rf node_modules dist

version-check: ## Check current version and suggest next
	@./scripts/version-check.sh

release: ## Create a new release (usage: make release VERSION=1.2.3)
ifdef VERSION
	@./scripts/release.sh $(VERSION)
else
	@echo "❌ Please specify VERSION: make release VERSION=1.2.3"
endif

security: ## Run security scan
	@echo "🔒 Running security scan..."
	gosec ./...

deps-update: ## Update dependencies
	@echo "📦 Updating dependencies..."
	go get -u ./...
	go mod tidy
	cd frontend && npm update

cross-build: ## Build for all platforms
	@echo "🌍 Building for all platforms..."
	./scripts/build.sh --prod --platform linux/amd64
	./scripts/build.sh --prod --platform linux/arm64  
	./scripts/build.sh --prod --platform darwin/amd64
	./scripts/build.sh --prod --platform darwin/arm64
	./scripts/build.sh --prod --platform windows/amd64

setup: install ## Initial project setup
	@echo "🚀 Setting up SysMind development environment..."
	@if ! command -v wails >/dev/null 2>&1; then \
		echo "Installing Wails CLI..."; \
		go install github.com/wailsapp/wails/v2/cmd/wails@latest; \
	fi
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	@if ! command -v gosec >/dev/null 2>&1; then \
		echo "Installing gosec..."; \
		go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest; \
	fi
	@echo "✅ Development environment ready!"

doctor: ## Check development environment
	@echo "🏥 Checking development environment..."
	@echo -n "Go: "
	@go version 2>/dev/null || echo "❌ Not installed"
	@echo -n "Node.js: "
	@node --version 2>/dev/null || echo "❌ Not installed" 
	@echo -n "Wails: "
	@wails version 2>/dev/null || echo "❌ Not installed"
	@echo -n "golangci-lint: "
	@golangci-lint version 2>/dev/null || echo "❌ Not installed"
	@echo -n "gosec: "
	@gosec --version 2>/dev/null || echo "❌ Not installed"
	@echo -n "Git: "
	@git --version 2>/dev/null || echo "❌ Not installed"