.PHONY: $(shell ls)

help: ## Show this help.
	@ sed -nEe '/@sed/!s/[^:#]*##\s*/ /p' $(MAKEFILE_LIST) | column -tl 2

lint: ## Run only allowlisted linters.
	@ golangci-lint run

lint-all: ## Run all non-blocklisted linters.
	@ golangci-lint run -c .golangci.all.yml

test: tmp ## Run unit tests.
	@ set -a \
		&& . ./env.test \
		&& go test -coverprofile=tmp/cover.out  ./...

tmp: ## Create a tmp directory.
	@ mkdir -p tmp
