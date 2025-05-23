.PHONY: $(shell ls)

help: ## Show this help.
	@ sed -nEe '/@sed/!s/[^:#]*##\s*/ /p' $(MAKEFILE_LIST) | column -tl 2

lint: ## Run linters.
	@ golangci-lint run

run: ## Run the application.
	@ go run cmd/flashcards/main.go

test: tmp ## Run unit tests.
	@ go test -coverprofile=tmp/cover.out  ./...

tmp: ## Create a tmp directory.
	@ mkdir -p tmp
