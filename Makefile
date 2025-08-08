.PHONY: tidy test vet lint build all

GO ?= go

all: tidy vet test

mod:
	$(GO) mod download

tidy:
	$(GO) mod tidy

vet:
	$(GO) vet ./...

test:
	$(GO) test -v ./...

build:
	$(GO) build ./...
