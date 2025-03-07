# Variables
APP_NAME := mcp
MAIN_PACKAGE := ./cmd/$(APP_NAME)
BUILD_DIR := build
DOCKER_IMAGE := $(APP_NAME)
DOCKER_TAG := latest
FRONTEND_DIR := mcp-frontend

VERSION ?= $(shell git describe --tags --always --dirty)
GIT_COMMIT ?= $(shell git rev-parse --short HEAD)
BUILD_TIME ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GO_VERSION = $(shell go version | cut -d ' ' -f 3)

# Enhanced build flags
LDFLAGS := -w -s \
	-X 'main.Version=$(VERSION)' \
	-X 'main.GitCommit=$(GIT_COMMIT)' \
	-X 'main.BuildTime=$(BUILD_TIME)' \
	-X 'main.GoVersion=$(GO_VERSION)'

# Go related variables
GOBASE := $(shell pwd)
GOBIN := $(GOBASE)/bin
GOFILES := $(shell find . -type f -name '*.go' -not -path "./vendor/*")

# Build flags
LDFLAGS := -w -s

.PHONY: all build clean test lint docker-build run help wire

all: clean build test

wire:
	@echo "Checking wire installation..."
	@which wire >/dev/null 2>&1 || (echo "Installing wire..." && go install github.com/google/wire/cmd/wire@latest)
	@echo "Generating wire_gen.go..."
	@cd cmd && wire
	@echo "Wire generation complete"

# Run the API server
run-api: wire
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
		go install github.com/air-verse/air@latest; \
		air -- api; \
	fi

# Development mode with hot reload for server
dev-server:
	@if command -v air >/dev/null; then \
		air -- server; \
	else \
		echo "Air is not installed. Installing air for hot reload..."; \
		go install github.com/air-verse/air@latest; \
		air -- server; \
	fi

# Run the MCP Kit Frontend
run-frontend:
	@echo "Running MCP Kit Frontend..."
	@if [ $$(docker ps -a -q -f name=mcp-frontend) ]; then \
		echo "Stopping and removing existing container..."; \
		docker stop mcp-frontend; \
		docker rm mcp-frontend; \
	fi
	@docker run -d \
		--name mcp-frontend \
		-p 3001:80 \
		-e VITE_MCP_BACKEND_API_ENDPOINT=http://localhost:8081 \
		ghcr.io/shaharia-lab/mcp-frontend:latest

# Run all components in development mode
dev-all:
	@echo "Starting all components in development mode..."
	@make -j 3 dev-api dev-server frontend-dev

# Build the application
build: wire
	@echo "Building optimized binary for $(APP_NAME)..."
	@CGO_ENABLED=0 go build \
		-trimpath \
		-ldflags="$(LDFLAGS)" \
		-a \
		-installsuffix cgo \
		-o $(BUILD_DIR)/$(APP_NAME)

# Build the application
build-in-docker: wire
	@echo "Building optimized binary for $(APP_NAME)..."
	@CGO_ENABLED=0 go build \
		-trimpath \
		-ldflags="$(LDFLAGS)" \
		-a \
		-installsuffix cgo \
		-o $(BUILD_DIR)/$(APP_NAME)

# Add a development build target (faster builds for development)
build-dev: wire
	@echo "Building development binary for $(APP_NAME)..."
	@go build -o $(BUILD_DIR)/$(APP_NAME)

# Load environment variables
load-env:
	@echo "Loading environment variables from .env.local..."
	@export $(grep -v '^#' .env.local | xargs)

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

# Run both frontend and backend in development mode
dev-all:
	@echo "Starting both frontend and backend in development mode..."
	@make -j 2 dev frontend-dev

# Build everything for production
build-all: build
	@echo "Building both frontend and backend for production..."

# Show help
help:
	@echo "Available targets:"
	@echo "  wire            - Generate wire_gen.go file"
	@echo "  build           - Build the backend application"
	@echo "  clean           - Clean build artifacts"
	@echo "  test            - Run tests"
	@echo "  lint            - Run linter"
	@echo "  docker-build    - Build docker image"
	@echo "  run             - Run the backend application"
	@echo "  dev             - Run backend in development mode with hot reload"
	@echo "  load-env		 - Load environment variables from .env.local"
	@echo "  run-frontend    - Run frontend docker"
	@echo "  dev-all        - Run both frontend and backend in development mode"
	@echo "  build-all      - Build both frontend and backend for production"
	@echo "  help            - Show this help message"

# Default target
.DEFAULT_GOAL := help