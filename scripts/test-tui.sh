#!/bin/bash
set -e

echo "=== TUI Application Testing ==="

# Setup dependencies first
echo "🔧 Setting up dependencies..."
go mod tidy
go mod download

# Run tests
echo "🧪 Running unit tests..."
go test ./... -v

# Build the application
echo "🏗️  Building TUI application..."
make build

# Test that the binary exists and can show help
echo "📋 Testing binary execution..."
if [ -f "./bin/archwsl-tui-configurator" ]; then
    echo "✅ Binary built successfully"

    # Test version flag
    ./bin/archwsl-tui-configurator --version
    echo "✅ Version flag works"

    # Test help flag
    ./bin/archwsl-tui-configurator --help
    echo "✅ Help flag works"

    echo ""
    echo "🎉 TUI application ready!"
    echo "📱 Run './bin/archwsl-tui-configurator' to start the TUI"
    echo "⌨️  Use Ctrl+C or 'q' to quit the application"
else
    echo "❌ Binary not found!"
    exit 1
fi

