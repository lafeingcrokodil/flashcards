.DEFAULT_GOAL := help
.PHONY: help web test cover vet lint vuln run build

help: ## List the targets.
	@grep -E '^[a-z]+:.*## ' $(MAKEFILE_LIST) | awk -F ':.*## ' '{printf "%-6s %s\n", $$1, $$2}'

# The Go binary embeds web/build, so the Go code only compiles if that
# directory exists. The targets that only check the Go code create a
# placeholder page if the front end hasn't been built, so that they don't need
# Node. The targets that build or run a binary build the front end instead.
FRONTEND := web/build/index.html

$(FRONTEND):
	mkdir -p web/build
	echo '<!doctype html><title>Flashcards</title><p>The front end has not been built. Run make web.</p>' > $@

web: ## Build the front end into web/build/.
	cd web && pnpm install --frozen-lockfile && pnpm build

test: $(FRONTEND) ## Run the tests.
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
cover: $(FRONTEND) ## Run the tests and check the coverage of each package in internal/.
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

vet: $(FRONTEND) ## Run go vet.
	go vet ./...

lint: $(FRONTEND) ## Run golangci-lint.
	golangci-lint run ./...

vuln: $(FRONTEND) ## Check for known vulnerabilities in the code and the Go version.
	go tool govulncheck ./...

run: web ## Build the front end and run the server, on port 8080 unless PORT is set.
	go run ./cmd/flashcards

build: web ## Build the front end and a server binary that includes it, into bin/.
	go build -o bin/flashcards ./cmd/flashcards
