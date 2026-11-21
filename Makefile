.DEFAULT_GOAL := help
.PHONY: help test cover vet lint run build

help: ## List the targets.
	@grep -E '^[a-z]+:.*## ' $(MAKEFILE_LIST) | awk -F ':.*## ' '{printf "%-6s %s\n", $$1, $$2}'

test: ## Run the tests.
	go test -race ./...

# The minimum percentage of statements that the tests must cover in each
# package in internal/. The other packages only wire things together, so they
# aren't checked.
MIN_COVERAGE := 80

# The module's path, which starts every package path in the coverage profile.
MODULE := $(shell go list -m)

# Each line of the profile after the first is "file:range statements count",
# where file is a path that starts with the module's path, and a file's
# package is its directory.
cover: ## Run the tests and check the coverage of each package in internal/.
	go test -race -coverprofile=coverage.out ./...
	@awk -v min=$(MIN_COVERAGE) -v module=$(MODULE) ' \
		NR > 1 { \
			pkg = $$1; \
			sub(/\/[^\/]*$$/, "", pkg); \
			pkg = substr(pkg, length(module) + 2); \
			if (pkg !~ /^internal\//) next; \
			total[pkg] += $$2; \
			if ($$3 > 0) covered[pkg] += $$2 \
		} \
		END { \
			for (pkg in total) { \
				pct = total[pkg] ? 100 * covered[pkg] / total[pkg] : 100; \
				printf "%s: %.1f%% (minimum %d%%)\n", pkg, pct, min; \
				if (pct < min) failed = 1 \
			} \
			exit failed \
		}' coverage.out

vet: ## Run go vet.
	go vet ./...

lint: ## Run golangci-lint.
	golangci-lint run ./...

run: ## Run the server locally, on port 8080 unless PORT is set.
	go run ./cmd/flashcards

build: ## Build the server into bin/.
	go build -o bin/flashcards ./cmd/flashcards
