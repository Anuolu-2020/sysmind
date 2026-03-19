# Version information
VERSION ?= $(shell git describe --tags --always 2>/dev/null || echo "dev")
GIT_COMMIT ?= $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
GIT_TAG ?= $(shell git describe --tags --exact-match 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
BUILD_USER ?= $(shell whoami)

# Build flags
LDFLAGS = -X 'sysmind/internal/version.Version=$(VERSION)' \
          -X 'sysmind/internal/version.GitCommit=$(GIT_COMMIT)' \
          -X 'sysmind/internal/version.GitTag=$(GIT_TAG)' \
          -X 'sysmind/internal/version.BuildDate=$(BUILD_DATE)' \
          -X 'sysmind/internal/version.BuildUser=$(BUILD_USER)'

.PHONY: help install dev build test clean lint release version-check install-linux install-user

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
	@echo "Version: $(VERSION)"
	@echo "Commit: $(shell echo $(GIT_COMMIT) | cut -c1-8)"
	./scripts/generate-desktop-file.sh .
	wails build -clean -upx -s -ldflags "$(LDFLAGS)"

build-dev: ## Build for development
	@echo "🏗️ Building for development..."
	@echo "Version: $(VERSION)"
	./scripts/generate-desktop-file.sh .
	wails build -ldflags "$(LDFLAGS)"

test: ## Run all tests
	@echo "🧪 Running tests..."
	cd frontend && npm run build
	go test -v -race ./...
	cd frontend && npm test -- --watchAll=false

test-coverage: ## Run tests with coverage
	@echo "📊 Running tests with coverage..."
	cd frontend && npm run build
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

version: ## Show version information
	@echo "SysMind Version Information:"
	@echo "  Version:    $(VERSION)"
	@echo "  Git Commit: $(shell echo $(GIT_COMMIT) | cut -c1-8)"
	@echo "  Git Tag:    $(GIT_TAG)"
	@echo "  Build Date: $(BUILD_DATE)"
	@echo "  Build User: $(BUILD_USER)"

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
	@echo "Version: $(VERSION)"
	./scripts/generate-desktop-file.sh .
	wails build -clean -platform linux/amd64 -upx -s -ldflags "$(LDFLAGS)"
	wails build -clean -platform linux/arm64 -upx -s -ldflags "$(LDFLAGS)"
	wails build -clean -platform darwin/amd64 -upx -s -ldflags "$(LDFLAGS)"
	wails build -clean -platform darwin/arm64 -upx -s -ldflags "$(LDFLAGS)"
	wails build -clean -platform windows/amd64 -upx -s -ldflags "$(LDFLAGS)"

install-user: build ## Install for current user (~/.local)
	@echo "📦 Installing SysMind for user..."
	@mkdir -p $$HOME/.local/bin
	@mkdir -p $$HOME/.local/share/applications
	@mkdir -p $$HOME/.local/share/pixmaps
	@cp build/bin/sysmind $$HOME/.local/bin/
	@cp build/icons/icon_512.png $$HOME/.local/share/pixmaps/sysmind.png
	@./scripts/generate-desktop-file.sh $$HOME/.local
	@cp sysmind.desktop $$HOME/.local/share/applications/
	@echo "✅ Installed to $$HOME/.local/"
	@echo "   Run 'export PATH=$$PATH:$$HOME/.local/bin' to add to PATH"

install-system: build ## Install system-wide (requires sudo)
	@echo "📦 Installing SysMind system-wide..."
	@echo "This requires sudo. You will be prompted for your password."
	sudo mkdir -p /usr/local/bin
	sudo mkdir -p /usr/local/share/applications
	sudo mkdir -p /usr/local/share/pixmaps
	sudo cp build/bin/sysmind /usr/local/bin/
	sudo cp build/icons/icon_512.png /usr/local/share/pixmaps/sysmind.png
	sudo ./scripts/generate-desktop-file.sh /usr/local
	sudo cp sysmind.desktop /usr/local/share/applications/
	sudo update-desktop-database /usr/local/share/applications 2>/dev/null || true
	@echo "✅ Installed to /usr/local/"

uninstall-user: ## Remove user installation
	@echo "🗑️  Uninstalling SysMind from user directory..."
	@rm -f $$HOME/.local/bin/sysmind
	@rm -f $$HOME/.local/share/pixmaps/sysmind.png
	@rm -f $$HOME/.local/share/applications/sysmind.desktop
	@echo "✅ User installation removed"

uninstall-system: ## Remove system installation (requires sudo)
	@echo "🗑️  Uninstalling SysMind from system..."
	sudo rm -f /usr/local/bin/sysmind
	sudo rm -f /usr/local/share/pixmaps/sysmind.png
	sudo rm -f /usr/local/share/applications/sysmind.desktop
	sudo update-desktop-database /usr/local/share/applications 2>/dev/null || true
	@echo "✅ System installation removed"

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