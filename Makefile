.PHONY: build test lint clean fmt vet

# Build the application
build:
	go build -o bin/archinstall-wsl ./cmd/archinstall-wsl

# Run tests
test:
	go test ./... -race

# Run linting
lint:
	gofmt -s .
	go vet ./...

# Clean build artifacts
clean:
	rm -rf bin/
	go clean

# Format code
fmt:
	gofmt -s -w .

# Run vet
vet:
	go vet ./...

# Install dependencies
deps:
	go mod download
	go mod tidy

# Run all checks
check: fmt vet test

