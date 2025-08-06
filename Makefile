# Makefile for archinstall-wsl

# Variables
BINARY_NAME=archwsl-tui-configurator
MAIN_PATH=./cmd/archinstall-wsl
BUILD_DIR=./bin

# Default target
.PHONY: all
all: clean lint test build

# Build the application
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	go test ./... -race

# Run linting
.PHONY: lint
lint:
	@echo "Running linting..."
	gofmt -s -w .
	go vet ./...

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)

# Run the application
.PHONY: run
run: build
	@echo "Running $(BINARY_NAME)..."
	$(BUILD_DIR)/$(BINARY_NAME)

# Install dependencies
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Check formatting
.PHONY: fmt-check
fmt-check:
	@echo "Checking formatting..."
	@test -z "$$(gofmt -l .)" || (echo "Files not formatted:" && gofmt -l . && exit 1)

# Display help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all       - Clean, lint, test, and build"
	@echo "  build     - Build the application"
	@echo "  test      - Run tests with race detection"
	@echo "  lint      - Run gofmt and go vet"
	@echo "  clean     - Clean build artifacts"
	@echo "  run       - Build and run the application"
	@echo "  deps      - Install and tidy dependencies"
	@echo "  fmt-check - Check if code is properly formatted"
	@echo "  help      - Display this help message"
