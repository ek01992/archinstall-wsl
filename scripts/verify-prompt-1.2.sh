#!/bin/bash
set -e

echo "=== Prompt 1.2 Verification: Basic TUI Application Shell ==="

# Required components verification
echo "🔍 Verifying TUI implementation..."

# 1. Check Bubble Tea dependency
echo "📦 Checking Bubble Tea dependency..."
if grep -q "github.com/charmbracelet/bubbletea" go.mod; then
    echo "✅ Bubble Tea dependency found"
else
    echo "❌ Bubble Tea dependency missing"
    exit 1
fi

# 2. Check main application entry point
echo "📍 Checking main application entry point..."
if [ -f "cmd/archinstall-wsl/main.go" ]; then
    if grep -q "tea.NewProgram" cmd/archinstall-wsl/main.go; then
        echo "✅ Main application uses Bubble Tea"
    else
        echo "❌ Main application doesn't use Bubble Tea"
        exit 1
    fi
else
    echo "❌ Main application file not found"
    exit 1
fi

# 3. Check multi-view model implementation
echo "🏗️  Checking multi-view model..."
if [ -f "internal/tui/model.go" ]; then
    if grep -q "WelcomeView\|MenuView" internal/tui/model.go; then
        echo "✅ Multi-view model implemented"
    else
        echo "❌ Multi-view model not found"
        exit 1
    fi
else
    echo "❌ TUI model file not found"
    exit 1
fi

# 4. Check Welcome view
echo "👋 Checking Welcome view..."
if [ -f "internal/tui/welcome.go" ]; then
    if grep -q "WelcomeModel\|View()" internal/tui/welcome.go; then
        echo "✅ Welcome view implemented"
    else
        echo "❌ Welcome view incomplete"
        exit 1
    fi
else
    echo "❌ Welcome view file not found"
    exit 1
fi

# 5. Check Menu view with placeholder tasks
echo "📋 Checking Menu view..."
if [ -f "internal/tui/menu.go" ]; then
    if grep -q "MenuModel\|Package Selection\|Network Configuration" internal/tui/menu.go; then
        echo "✅ Menu view with placeholder tasks implemented"
    else
        echo "❌ Menu view incomplete or missing tasks"
        exit 1
    fi
else
    echo "❌ Menu view file not found"
    exit 1
fi

# 6. Check clean shutdown on ctrl+c
echo "🛑 Checking clean shutdown implementation..."
if grep -q "ctrl+c.*tea.Quit\|tea.KeyCtrlC" internal/tui/model.go; then
    echo "✅ Clean shutdown on Ctrl+C implemented"
else
    echo "❌ Clean shutdown not found"
    exit 1
fi

# 7. Build and test the application
echo "🏗️  Building application..."
go mod tidy
make build

if [ -f "./bin/archwsl-tui-configurator" ]; then
    echo "✅ Application builds successfully"

    # Test command line flags still work
    ./bin/archwsl-tui-configurator --version > /dev/null
    echo "✅ Command line flags work"

    echo ""
    echo "🎉 Prompt 1.2 Verification PASSED!"
    echo ""
    echo "✅ Requirements Met:"
    echo "   • Bubble Tea library integrated"
    echo "   • Main application entry point in /cmd/archinstall-wsl"
    echo "   • Multi-view model implemented"
    echo "   • Welcome view created"
    echo "   • Main menu view with placeholder tasks"
    echo "   • Clean shutdown on Ctrl+C"
    echo "   • Application compiles and runs"
    echo ""
    echo "🚀 Ready for verification: Run './bin/archwsl-tui-configurator' to test the TUI!"
else
    echo "❌ Application build failed"
    exit 1
fi

