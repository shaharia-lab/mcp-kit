# Variables
APP_NAME := mcp
MAIN_PACKAGE := ./cmd/$(APP_NAME)
BUILD_DIR := build
DOCKER_IMAGE := $(APP_NAME)
DOCKER_TAG := latest
FRONTEND_DIR := mcp-frontend

# Go related variables
GOBASE := $(shell pwd)
GOBIN := $(GOBASE)/bin
GOFILES := $(shell find . -type f -name '*.go' -not -path "./vendor/*")

# Build flags
LDFLAGS := -w -s

.PHONY: all build clean test lint docker-build run help frontend-dev frontend-build frontend-install

all: clean build test

# Run the API server
run-api:
	@echo "Running API server..."
	@go run . api

# Run the main server
run-server:
	@echo "Running main server..."
	@go run . server

# Development mode with hot reload for API
dev-api:
	@if command -v air >/dev/null; then \
		air -- api; \
	else \
		echo "Air is not installed. Installing air for hot reload..."; \
		go install github.com/cosmtrek/air@latest; \
		air -- api; \
	fi

# Development mode with hot reload for server
dev-server:
	@if command -v air >/dev/null; then \
		air -- server; \
	else \
		echo "Air is not installed. Installing air for hot reload..."; \
		go install github.com/cosmtrek/air@latest; \
		air -- server; \
	fi

# Run all components in development mode
dev-all:
	@echo "Starting all components in development mode..."
	@make -j 3 dev-api dev-server frontend-dev

# Build the application
build:
	@echo "Building $(APP_NAME)..."
	@CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME)

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@go clean -testcache
	@echo "Cleaning frontend build..."
	@rm -rf $(FRONTEND_DIR)/dist

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run linter
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint is not installed. Please install it first."; \
		exit 1; \
	fi

# Build docker image
docker-build:
	@echo "Building docker image..."
	@docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

# Run the backend application
run:
	@echo "Running $(APP_NAME)..."
	@go run $(MAIN_PACKAGE)

# Development mode with hot reload for backend
dev:
	@if command -v air >/dev/null; then \
		air; \
	else \
		echo "Air is not installed. Installing air for hot reload..."; \
		go install github.com/cosmtrek/air@latest; \
		air; \
	fi

# Install frontend dependencies
frontend-install:
	@echo "Installing frontend dependencies..."
	@cd $(FRONTEND_DIR) && npm install

# Run frontend in development mode
frontend-dev:
	@echo "Starting frontend development server..."
	@cd $(FRONTEND_DIR) && npm run dev

# Build frontend for production
frontend-build:
	@echo "Building frontend for production..."
	@cd $(FRONTEND_DIR) && npm run build

# Run both frontend and backend in development mode
dev-all:
	@echo "Starting both frontend and backend in development mode..."
	@make -j 2 dev frontend-dev

# Build everything for production
build-all: build frontend-build
	@echo "Building both frontend and backend for production..."

# Show help
help:
	@echo "Available targets:"
	@echo "  build           - Build the backend application"
	@echo "  clean           - Clean build artifacts"
	@echo "  test            - Run tests"
	@echo "  lint            - Run linter"
	@echo "  docker-build    - Build docker image"
	@echo "  run             - Run the backend application"
	@echo "  dev             - Run backend in development mode with hot reload"
	@echo "  frontend-install- Install frontend dependencies"
	@echo "  frontend-dev    - Run frontend in development mode"
	@echo "  frontend-build  - Build frontend for production"
	@echo "  dev-all        - Run both frontend and backend in development mode"
	@echo "  build-all      - Build both frontend and backend for production"
	@echo "  help            - Show this help message"

# Default target
.DEFAULT_GOAL := help