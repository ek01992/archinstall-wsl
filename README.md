# ArchInstall WSL

[![CI](https://github.com/ekowald/archinstall-wsl/actions/workflows/ci.yml/badge.svg)](https://github.com/ekowald/archinstall-wsl/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/ekowald/archinstall-wsl/branch/main/graph/badge.svg)](https://codecov.io/gh/ekowald/archinstall-wsl)

A modern Terminal User Interface (TUI) configurator for ArchInstall on WSL environments. Built with Go and Bubble Tea for an intuitive command-line experience.

## Project Structure

```text
archinstall-wsl/
├── cmd/
│   └── archinstall-wsl/     # Main application entry point
│       ├── main.go
│       └── main_test.go
├── internal/                # Private application and library code
├── pkg/                     # Library code that can be used by external applications
├── .github/
│   └── workflows/
│       └── ci.yml          # GitHub Actions CI pipeline
├── Makefile                # Build automation
├── go.mod                  # Go module definition
└── README.md
```

## Getting Started

### Prerequisites

- Go 1.22.2 or later
- Make (optional, but recommended)

### Building

```bash
# Using Make
make build

# Or directly with Go
go build -o bin/archwsl-tui-configurator ./cmd/archinstall-wsl
```

### Running

```bash
# Using Make
make run

# Or directly
./bin/archwsl-tui-configurator

# To test the TUI application
./scripts/test-tui.sh
```

### TUI Features

- **Welcome Screen**: Beautiful introduction with auto-transition
- **Main Menu**: Interactive menu with ArchInstall configuration tasks
- **Keyboard Navigation**: Vi-style (hjkl) and arrow key support
- **Clean Shutdown**: Ctrl+C or 'q' to quit gracefully
- **Responsive Design**: Adapts to terminal size

### Development

```bash
# Install dependencies
make deps

# Run tests
make test

# Run linting
make lint

# Clean build artifacts
make clean

# See all available targets
make help
```

## CI/CD

The project uses GitHub Actions for continuous integration. The CI pipeline:

- Runs on Go 1.22.2
- Checks code formatting with `gofmt -s`
- Runs static analysis with `go vet`
- Executes tests with race detection
- Builds the binary
- Tests basic command-line functionality

## Contributing

1. Ensure your code is properly formatted: `make lint`
2. Run tests: `make test`
3. Build to verify: `make build`

## License

[Add your license here]
