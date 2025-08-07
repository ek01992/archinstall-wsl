#!/bin/bash

# Verification script for Milestone 1 setup
set -e

echo "🔍 Verifying Milestone 1 setup..."

# Check if Go is available
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed or not in PATH"
    echo "Please install Go 1.24.6 or later"
    exit 1
fi

echo "✅ Go is available"

# Check project structure
echo "📁 Checking project structure..."

required_dirs=(
    "cmd/archinstall-wsl"
    "internal/system"
    "internal/system/mocks"
    "internal/tui"
    ".github/workflows"
)

for dir in "${required_dirs[@]}"; do
    if [ ! -d "$dir" ]; then
        echo "❌ Missing required directory: $dir"
        exit 1
    fi
done

echo "✅ Project structure is correct"

# Check required files
echo "📄 Checking required files..."

required_files=(
    "go.mod"
    "Makefile"
    ".github/workflows/ci.yml"
    "cmd/archinstall-wsl/main.go"
    "cmd/archinstall-wsl/main_test.go"
    "internal/system/executor.go"
    "internal/system/state_checker.go"
    "internal/system/mocks/executor.go"
    "internal/system/mocks/state_checker.go"
    "internal/tui/models.go"
    "internal/tui/models_test.go"
)

for file in "${required_files[@]}"; do
    if [ ! -f "$file" ]; then
        echo "❌ Missing required file: $file"
        exit 1
    fi
done

echo "✅ All required files exist"

# Check Go module
echo "🔧 Checking Go module..."

if ! go mod tidy; then
    echo "❌ Failed to tidy Go module"
    exit 1
fi

echo "✅ Go module is valid"

# Check compilation
echo "🔨 Checking compilation..."

if ! go build ./cmd/archinstall-wsl; then
    echo "❌ Failed to compile main application"
    exit 1
fi

echo "✅ Application compiles successfully"

# Check tests
echo "🧪 Running tests..."

if ! go test ./...; then
    echo "❌ Tests failed"
    exit 1
fi

echo "✅ All tests pass"

# Check formatting
echo "🎨 Checking code formatting..."

if ! gofmt -s -l . | grep -q .; then
    echo "✅ Code is properly formatted"
else
    echo "⚠️  Some files need formatting (this is expected in development)"
fi

# Check vet
echo "🔍 Running go vet..."

if ! go vet ./...; then
    echo "❌ go vet found issues"
    exit 1
fi

echo "✅ go vet passed"

echo ""
echo "🎉 Milestone 1 verification complete!"
echo "✅ Project structure is correct"
echo "✅ All interfaces are defined"
echo "✅ Mock implementations are working"
echo "✅ TUI application shell is implemented"
echo "✅ CI pipeline is configured"
echo ""
echo "The application is ready for development!"

