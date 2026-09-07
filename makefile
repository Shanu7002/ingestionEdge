# Define the default target when someone just runs 'make'
.DEFAULT_GOAL := all
DB_CONTAINER := telemetry-db
DB_USER := secretUser
DB_NAME := edge_ingestion
CHANGELOG := internal/migrations/changelog.xml
MIGRATIONS_DIR := internal/migrations

.PHONY: setup fmt lint test build migrate all

all: fmt lint test migrate build

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

migrate:
	@python3 -c '\
	import xml.etree.ElementTree as ET; \
	root = ET.parse("$(CHANGELOG)").getroot(); \
	print("\n".join(m.find("file").text for m in root.findall("migration"))); \
	' | while read file; do \
		echo "Running migration: $$file"; \
		docker exec -i $(DB_CONTAINER) psql \
			-U $(DB_USER) \
			-d $(DB_NAME) \
			-v ON_ERROR_STOP=1 \
			< $(MIGRATIONS_DIR)/$$file || exit 1; \
	done
