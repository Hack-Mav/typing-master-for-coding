.PHONY: help install dev build test clean lint format docker-up docker-down

# Default target
help:
	@echo "Typing Master for Coding - Development Commands"
	@echo ""
	@echo "Installation:"
	@echo "  install     Install dependencies for all applications"
	@echo ""
	@echo "Development:"
	@echo "  dev         Start all development services"
	@echo "  dev-web     Start web frontend only"
	@echo "  dev-api     Start API backend only"
	@echo ""
	@echo "Building:"
	@echo "  build       Build all applications"
	@echo "  build-web   Build web frontend only"
	@echo "  build-api   Build API backend only"
	@echo ""
	@echo "Testing:"
	@echo "  test        Run all tests"
	@echo "  test-web    Run web frontend tests"
	@echo "  test-api    Run API backend tests"
	@echo "  test-e2e    Run end-to-end tests"
	@echo ""
	@echo "Code Quality:"
	@echo "  lint        Run linting for all applications"
	@echo "  format      Format code for all applications"
	@echo ""
	@echo "Docker:"
	@echo "  docker-up   Start development containers"
	@echo "  docker-down Stop development containers"
	@echo "  clean       Clean build artifacts and dependencies"

# Installation
install:
	@echo "Installing dependencies..."
	cd apps/web && npm install
	cd apps/api && go mod download

# Development
dev: docker-up
	@echo "Starting all development services..."
	@echo "Web: http://localhost:3000"
	@echo "API: http://localhost:8080"

dev-web:
	@echo "Starting web frontend..."
	cd apps/web && npm start

dev-api:
	@echo "Starting API backend..."
	cd apps/api && go run cmd/server/main.go

# Building
build: build-web build-api

build-web:
	@echo "Building web frontend..."
	cd apps/web && npm run build

build-api:
	@echo "Building API backend..."
	cd apps/api && go build -o bin/server cmd/server/main.go

# Testing
test: test-web test-api

test-web:
	@echo "Running web frontend tests..."
	cd apps/web && npm test -- --coverage --watchAll=false

test-api:
	@echo "Running API backend tests..."
	cd apps/api && go test -v -race -coverprofile=coverage.out ./...

test-e2e:
	@echo "Running end-to-end tests..."
	cd tests/e2e && npm test

# Code Quality
lint: lint-web lint-api

lint-web:
	@echo "Linting web frontend..."
	cd apps/web && npm run lint

lint-api:
	@echo "Linting API backend..."
	cd apps/api && go vet ./...

format: format-web format-api

format-web:
	@echo "Formatting web frontend..."
	cd apps/web && npm run format

format-api:
	@echo "Formatting API backend..."
	cd apps/api && go fmt ./...

# Docker
docker-up:
	@echo "Starting development containers..."
	docker-compose up -d postgres redis

docker-down:
	@echo "Stopping development containers..."
	docker-compose down

# Cleanup
clean:
	@echo "Cleaning build artifacts..."
	rm -rf apps/web/build
	rm -rf apps/web/node_modules
	rm -rf apps/api/bin
	cd apps/api && go clean -cache
	docker-compose down -v
