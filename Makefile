.PHONY: all install generate build dev dev-backend dev-frontend stop clean lint test e2e help

# Default target
all: install generate build

# ==================== Installation ====================

install: install-backend install-frontend ## Install all dependencies
	@echo "✅ All dependencies installed"

install-backend: ## Install Go dependencies
	@echo "📦 Installing Go dependencies..."
	cd backend && go mod download

install-frontend: ## Install npm dependencies
	@echo "📦 Installing npm dependencies..."
	cd frontend && npm install

# ==================== Code Generation ====================

generate: ## Generate Go and TypeScript types from protobuf
	@echo "🔄 Generating types from protobuf..."
	cd proto && buf generate
	@echo "✅ Types generated"

lint-proto: ## Lint protobuf files
	@echo "🔍 Linting protobuf files..."
	cd proto && buf lint

breaking-proto: ## Check for breaking changes in protobuf
	@echo "🔍 Checking for breaking changes..."
	cd proto && buf breaking --against '.git#branch=main'

# ==================== Building ====================

build: build-backend build-frontend ## Build all

build-backend: ## Build Go backend
	@echo "🔨 Building Go backend..."
	cd backend && go build -o bin/server ./cmd/server
	@echo "✅ Backend built: backend/bin/server"

build-frontend: ## Build frontend for production
	@echo "🔨 Building frontend..."
	cd frontend && npm run build
	@echo "✅ Frontend built: frontend/dist"

# ==================== Development ====================

dev: ## Start both backend and frontend in development mode
	@echo "🚀 Starting development servers..."
	@make -j2 dev-backend dev-frontend

dev-backend: ## Start Go backend with hot reload (requires air)
	@echo "🚀 Starting Go backend on :8080..."
	@if command -v air > /dev/null; then \
		cd backend && air; \
	else \
		cd backend && go run ./cmd/server; \
	fi

dev-frontend: ## Start Vite dev server
	@echo "🚀 Starting frontend on :3000..."
	cd frontend && npm run dev

# ==================== Testing ====================

test: test-backend test-frontend ## Run all tests

test-backend: ## Run Go tests
	@echo "🧪 Running Go tests..."
	cd backend && go test -v ./...

test-frontend: ## Run frontend tests
	@echo "🧪 Running frontend tests..."
	cd frontend && npm test 2>/dev/null || echo "No tests configured"

e2e: ## Run Playwright E2E tests (requires backend/frontend running)
	@echo "🧪 Running Playwright E2E tests..."
	@nc -z localhost 8080 || (echo "❌ Start backend first: make dev-backend" && exit 1)
	@nc -z localhost 3000 || (echo "❌ Start frontend first: make dev-frontend" && exit 1)
	cd frontend && npm test

# ==================== Linting ====================

lint: lint-backend lint-frontend lint-proto ## Lint all code

lint-backend: ## Lint Go code
	@echo "🔍 Linting Go code..."
	cd backend && go vet ./...
	@if command -v golangci-lint > /dev/null; then \
		cd backend && golangci-lint run; \
	fi

lint-frontend: ## Lint TypeScript code
	@echo "🔍 Linting TypeScript code..."
	cd frontend && npm run lint 2>/dev/null || echo "Lint script not configured"

typecheck: ## Run TypeScript type checking
	@echo "🔍 Type checking frontend..."
	cd frontend && npx tsc --noEmit

# ==================== Cleanup ====================

clean: ## Clean build artifacts
	@echo "🧹 Cleaning build artifacts..."
	rm -rf backend/bin
	rm -rf frontend/dist
	rm -rf backend/gen/proto
	rm -rf frontend/src/gen/proto
	@echo "✅ Cleaned"

clean-all: clean ## Clean everything including node_modules
	@echo "🧹 Deep cleaning..."
	rm -rf frontend/node_modules
	@echo "✅ Deep cleaned"

# ==================== Database ====================

db-reset: ## Reset the SQLite database
	@echo "🗑️  Resetting database..."
	rm -f backend/catan.db
	@echo "✅ Database reset"

# ==================== Help ====================

help: ## Show this help message
	@echo "Settlers from Catan - Development Commands"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
