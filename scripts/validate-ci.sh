#!/bin/bash
set -e

echo "=== ArchInstall WSL CI Validation ==="

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed"
    exit 1
fi

echo "✅ Go is installed: $(go version)"

# Check Go module
echo "🔍 Verifying Go module..."
go mod verify
echo "✅ Go module verified"

# Check formatting
echo "🔍 Checking code formatting..."
if [ "$(gofmt -s -l . | wc -l)" -gt 0 ]; then
    echo "❌ Code is not formatted. Run: gofmt -s -w ."
    gofmt -s -l .
    exit 1
fi
echo "✅ Code formatting is correct"

# Run go vet
echo "🔍 Running go vet..."
go vet ./...
echo "✅ go vet passed"

# Run tests
echo "🔍 Running tests with race detection..."
go test ./... -race
echo "✅ Tests passed"

# Test build
echo "🔍 Testing build..."
make build
echo "✅ Build successful"

# Test binary execution
echo "🔍 Testing binary execution..."
./bin/archwsl-tui-configurator --help
./bin/archwsl-tui-configurator --version

echo ""
echo "🎉 All CI validation checks passed!"
echo "The project is ready for GitHub Actions CI pipeline."
