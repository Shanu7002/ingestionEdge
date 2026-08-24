# Define the default target when someone just runs 'make'
.DEFAULT_GOAL := all
.PHONY: setup fmt lint test build all

all: fmt lint test build

setup:
	@echo "==> Configuring git hooks..."
	git config core.hooksPath .githooks
	@echo "==> Installing dependencies..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	@echo "==> Environment ready."

fmt:
	@echo "==> Formatting code and resolving imports..."
	goimports -w .

lint:
	@echo "==> Running golangci-lint..."
	golangci-lint run ./...

test:
	@echo "==> Running tests with race detector..."
	go test -v -race -cover ./...

build:
	@echo "==> Building binary..."
	go build -o bin/server cmd/main.go