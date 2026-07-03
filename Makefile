.PHONY: help build run clean test deps install lint

help:
	@echo "Photo Library - Available Commands"
	@echo "===================================="
	@echo "make deps       - Download Go dependencies"
	@echo "make build      - Build the application"
	@echo "make run        - Run the application"
	@echo "make dev        - Run in development mode with auto-reload"
	@echo "make test       - Run tests"
	@echo "make lint       - Run linter"
	@echo "make clean      - Clean build artifacts"
	@echo "make install    - Download dependencies and build"

deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

build: deps
	@echo "Building application..."
	go build -o photo-library

run: build
	@echo "Starting application on http://localhost:8082"
	./photo-library

dev: deps
	@echo "Running in development mode..."
	@if ! command -v air &> /dev/null; then \
		echo "Installing air for hot-reload..."; \
		go install github.com/cosmtrek/air@latest; \
	fi
	air

test:
	@echo "Running tests..."
	go test ./... -v

lint:
	@echo "Running linter..."
	@if ! command -v golangci-lint &> /dev/null; then \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	golangci-lint run

clean:
	@echo "Cleaning build artifacts..."
	rm -f photo-library
	rm -f photo_library.db
	go clean

install: deps build

.DEFAULT_GOAL := help
