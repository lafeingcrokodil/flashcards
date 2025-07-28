.PHONY: $(shell ls)

help: ## Show this help.
	@ sed -nEe '/@sed/!s/[^:#]*##\s*/ /p' $(MAKEFILE_LIST) | column -tl 2

lint: ## Run standard linters and a few additional explicitly enabled ones.
	@ golangci-lint run

lint-all: ## Run all linters that aren't explicitly disabled.
	@ golangci-lint run -c .golangci.all.yml

test: tmp ## Run unit tests.
	@ set -a \
		&& . ./env.test \
		&& go test -coverprofile=tmp/cover.out  ./...

tmp: ## Create a tmp directory.
	@ mkdir -p tmp
