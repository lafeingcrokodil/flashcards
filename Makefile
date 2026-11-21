.DEFAULT_GOAL := help
.PHONY: help test vet lint run build

help: ## List the targets.
	@grep -E '^[a-z]+:.*## ' $(MAKEFILE_LIST) | awk -F ':.*## ' '{printf "%-6s %s\n", $$1, $$2}'

test: ## Run the tests.
	go test -race ./...

vet: ## Run go vet.
	go vet ./...

lint: ## Run golangci-lint.
	golangci-lint run ./...

run: ## Run the server locally, on port 8080 unless PORT is set.
	go run ./cmd/flashcards

build: ## Build the server into bin/.
	go build -o bin/flashcards ./cmd/flashcards
