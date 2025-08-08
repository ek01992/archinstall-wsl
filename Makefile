.PHONY: tidy test vet lint build all check-go

GO ?= go
GOVERSION ?= 1.24.6

all: tidy lint vet test

mod:
	$(GO) mod download

check-go:
	@v=$$($(GO) version | awk '{print $$3}'); \
	if [ "$$v" != "go$(GOVERSION)" ]; then \
		echo "WARNING: Go version $$v != go$(GOVERSION). Recommended: $(GOVERSION)"; \
	fi

tidy:
	$(GO) mod tidy

lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest && \
		$$($(GO) env GOPATH)/bin/golangci-lint run ./...; \
	fi

vet:
	$(GO) vet ./...

test:
	$(GO) test -v ./...

build:
	$(GO) build ./...
