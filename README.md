# archinstall-wsl

A WSL Arch Linux installer with a terminal user interface (TUI).

## Project Structure

```md
.
├── cmd/archinstall-wsl/     # Main application entry point
├── internal/                 # Private application code
├── pkg/                     # Public packages
├── .github/workflows/       # CI/CD workflows
└── Makefile                 # Build and development tasks
```

## Development

### Prerequisites

- Go 1.24.6 or later

### Building

```bash
make build
```

### Testing

```bash
make test
```

### Linting

```bash
make lint
```

### Running all checks

```bash
make check
```

## CI/CD

The project uses GitHub Actions for continuous integration. The CI pipeline:

- Runs on push to main/develop branches and pull requests
- Uses Go 1.24.6
- Runs `gofmt -s .` for code formatting
- Runs `go vet ./...` for static analysis
- Runs `go test ./... -race` for tests with race detection
- Builds the application
