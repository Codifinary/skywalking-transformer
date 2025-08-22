.PHONY: help build run test clean docker-build docker-run docker-stop docker-push

# Default target
help:
	@echo "Available commands:"
	@echo "  build         - Build the Go binary"
	@echo "  run           - Run the application locally"
	@echo "  test          - Run tests"
	@echo "  clean         - Clean build artifacts"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-run    - Run with Docker Compose"
	@echo "  docker-stop   - Stop Docker Compose services"
	@echo "  docker-push   - Push Docker image to registry"

# Build the Go binary
build:
	@echo "Building binary..."
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o codexray-transformer .

# Run locally
run: build
	@echo "Running application..."
	./codexray-transformer

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f codexray-transformer
	rm -f *.log

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t skywalking-transformer:latest .

# Run with Docker Compose
docker-run:
	@echo "Starting services with Docker Compose..."
	docker-compose up -d

# Stop Docker Compose services
docker-stop:
	@echo "Stopping Docker Compose services..."
	docker-compose down

# Push Docker image to registry (requires login)
docker-push:
	@echo "Pushing Docker image to registry..."
	docker tag skywalking-transformer:latest ghcr.io/$(shell git config --get remote.origin.url | sed 's/.*github.com[:/]\([^/]*\/[^/]*\).*/\1/')
	docker push ghcr.io/$(shell git config --get remote.origin.url | sed 's/.*github.com[:/]\([^/]*\/[^/]*\).*/\1/')

# Development mode with hot reload (requires air)
dev:
	@echo "Starting development mode with hot reload..."
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "Air not found. Install with: go install github.com/cosmtrek/air@latest"; \
		echo "Or run: make run"; \
	fi
