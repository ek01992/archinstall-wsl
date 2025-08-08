.PHONY: tidy test vet lint build all

GO ?= go

all: tidy lint vet test

mod:
	$(GO) mod download

tidy:
	$(GO) mod tidy

lint:
	golangci-lint run ./...

vet:
	$(GO) vet ./...

test:
	$(GO) test -v ./...

build:
	$(GO) build ./...
