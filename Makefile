.PHONY: build run test clean swagger swagger-validate swagger-full dev migrate-help migrate-up migrate-down migrate-reset migrate-status migrate-list migrate-create migrate-force migrate-goto migrate-steps migrate-fix-dirty security-check safe-run safe-dev fmt lint check

build:
	go build -o bin/server cmd/server/main.go

run:
	go run cmd/server/main.go

# Security-aware run commands
safe-run: security-check
	go run cmd/server/main.go

safe-dev: security-check swagger
	go run cmd/server/main.go

# Security validation
security-check:
	@echo "Running security pre-flight checks..."
	@chmod +x scripts/security-check.sh
	@set -a && [ -f .env ] && . ./.env && set +a && scripts/security-check.sh

test:
	go test -v ./...

clean:
	rm -rf bin/ docs/

swagger:
	@echo "🔄 Generating OpenAPI 3.0 documentation..."
	$(HOME)/go/bin/swag init \
		-g cmd/server/main.go \
		-o docs \
		--parseInternal \
		--parseDependency \
		--parseDependencyLevel 3 \
		--packagePrefix linke/internal \
		--outputTypes go,json,yaml
	@echo "✅ Swagger documentation generated successfully"

swagger-validate:
	@echo "🔍 Validating OpenAPI specification..."
	@if command -v swagger >/dev/null 2>&1; then \
		swagger validate docs/swagger.json && echo "✅ OpenAPI specification is valid"; \
	elif command -v npx >/dev/null 2>&1; then \
		npx @apidevtools/swagger-cli validate docs/swagger.json && echo "✅ OpenAPI specification is valid"; \
	else \
		echo "⚠️  No validation tool found. Install 'swagger' or 'npx' for validation."; \
	fi

swagger-full: swagger swagger-validate

dev: swagger
	go run cmd/server/main.go

install:
	go mod tidy
	go install github.com/swaggo/swag/cmd/swag@latest
	go install mvdan.cc/gofumpt@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Database migration commands
# Uses integrated migration functionality in main server application

migrate-help:
	@echo "Showing migration help..."
	go run cmd/server/main.go -migrate-help

migrate-up:
	@echo "Running all pending migrations..."
	go run cmd/server/main.go -migrate-command=up

migrate-down:
	@echo "Rolling back one migration..."
	go run cmd/server/main.go -migrate-command=down

migrate-reset:
	@echo "Resetting database (drop all tables and re-run migrations)..."
	go run cmd/server/main.go -migrate-command=reset

migrate-status:
	@echo "Checking migration status..."
	go run cmd/server/main.go -migrate-command=status

migrate-list:
	@echo "Listing applied migrations..."
	go run cmd/server/main.go -migrate-command=list

migrate-create:
	@if [ -z "$(NAME)" ]; then \
		echo "ERROR: NAME variable is not set"; \
		echo "Usage: make migrate-create NAME=your_migration_name"; \
		exit 1; \
	fi
	go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	migrate create -ext sql -dir migrations -seq $(NAME)

migrate-force:
	@if [ -z "$(VERSION)" ]; then \
		echo "ERROR: VERSION variable is not set"; \
		echo "Usage: make migrate-force VERSION=N"; \
		exit 1; \
	fi
	@echo "Forcing migration version to $(VERSION)..."
	go run cmd/server/main.go -migrate-command=force -migrate-version=$(VERSION)

migrate-goto:
	@if [ -z "$(VERSION)" ]; then \
		echo "ERROR: VERSION variable is not set"; \
		echo "Usage: make migrate-goto VERSION=N"; \
		exit 1; \
	fi
	@echo "Migrating to version $(VERSION)..."
	go run cmd/server/main.go -migrate-command=goto -migrate-version=$(VERSION)

migrate-steps:
	@if [ -z "$(STEPS)" ]; then \
		echo "ERROR: STEPS variable is not set"; \
		echo "Usage: make migrate-steps STEPS=N (positive for up, negative for down)"; \
		exit 1; \
	fi
	@echo "Running $(STEPS) migration steps..."
	go run cmd/server/main.go -migrate-command=steps -migrate-steps=$(STEPS)

migrate-fix-dirty:
	@if [ -z "$(VERSION)" ]; then \
		echo "ERROR: VERSION variable is not set"; \
		echo "Usage: make migrate-fix-dirty VERSION=N"; \
		echo "Use the version number from the 'Dirty database version X' error"; \
		exit 1; \
	fi
	@echo "Fixing dirty migration state for version $(VERSION)..."
	go run cmd/server/main.go -migrate-command=fix-dirty -migrate-version=$(VERSION)

help:
	@echo "Available commands:"
	@echo "  build         - Build the application"
	@echo "  run           - Run the application"
	@echo "  test          - Run tests"
	@echo "  clean         - Clean build artifacts"
	@echo "  swagger       - Generate OpenAPI 3.0 documentation"
	@echo "  swagger-validate - Validate generated OpenAPI specification"
	@echo "  swagger-full  - Generate and validate OpenAPI documentation"
	@echo "  dev           - Run in development mode with swagger"
	@echo "  install       - Install dependencies and tools"
	@echo ""
	@echo "Database migration commands (integrated with main server):"
	@echo "  migrate-help        - Show detailed migration help"
	@echo "  migrate-up          - Run all pending migrations"
	@echo "  migrate-down        - Rollback one migration"
	@echo "  migrate-reset       - Drop all tables and re-run migrations"
	@echo "  migrate-status      - Show current migration version"
	@echo "  migrate-list        - List all applied migrations"
	@echo "  migrate-create NAME=name     - Create new migration files"
	@echo "  migrate-force VERSION=N      - Force set migration version"
	@echo "  migrate-goto VERSION=N       - Migrate to specific version"
	@echo "  migrate-steps STEPS=N        - Run N migration steps"
	@echo "  migrate-fix-dirty VERSION=N  - Fix dirty migration state"
	@echo ""
	@echo "Migration commands use environment variables from .env file:"
	@echo "  DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME"
	@echo ""
	@echo "Example usage:"
	@echo "  make migrate-up                    # Run all migrations"
	@echo "  make migrate-status                # Check current version"
	@echo "  make migrate-create NAME=add_users # Create new migration"
	@echo "  make migrate-goto VERSION=3        # Migrate to version 3"
	@echo "  make migrate-steps STEPS=2         # Run 2 migrations up"
	@echo "  make migrate-steps STEPS=-1        # Rollback 1 migration"

# Code formatting and linting
fmt:
	@echo "Formatting code with gofumpt..."
	$(HOME)/go/bin/gofumpt -w -extra .

lint:
	@echo "Running linter..."
	$(HOME)/go/bin/golangci-lint run

check: fmt lint
	@echo "Running all code quality checks..."
	go vet ./...
	go test -race ./...