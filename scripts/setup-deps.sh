#!/bin/bash
set -e

echo "=== Setting up Go dependencies ==="

# Download and verify dependencies
echo "📦 Downloading Go modules..."
go mod download

echo "🔧 Tidying Go modules..."
go mod tidy

echo "✅ Dependencies setup complete!"

# Verify module integrity
echo "🔍 Verifying module integrity..."
go mod verify

echo "🎉 All dependencies are ready!"

