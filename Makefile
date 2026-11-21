.DEFAULT_GOAL := help
.PHONY: help test lint run build

help: ## List the targets.
	@grep -E '^[a-z]+:.*## ' $(MAKEFILE_LIST) | awk -F ':.*## ' '{printf "%-6s %s\n", $$1, $$2}'

test: ## Run the tests.
	go test -race ./...

lint: ## Run go vet and golangci-lint.
	go vet ./...
	golangci-lint run

run: ## Run the server locally.
	go run ./cmd/flashcards

build: ## Build the server into bin/.
	go build -o bin/flashcards ./cmd/flashcards
