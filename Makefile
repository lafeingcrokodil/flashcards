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

tsc: ## Compile TypeScript code.
	@ . $(HOME)/.nvm/nvm.sh \
		&& nvm install stable \
		&& npm install \
		&& npx tsc

web-docker: tsc ## Run the application in a local Docker container.
	@ docker build -t flashcards .
	@ docker run --rm \
		-v $(HOME)/.config/gcloud:/root/.config/gcloud \
		--env-file env.production \
		-p 8080:8080 \
		flashcards

web-dev: tsc ## Run the application locally without using Docker.
	@ set -a \
		&& . ./env.production \
		&& go run cmd/web/main.go
