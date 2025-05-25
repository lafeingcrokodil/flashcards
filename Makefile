.PHONY: $(shell ls)

help: ## Show this help.
	@ sed -nEe '/@sed/!s/[^:#]*##\s*/ /p' $(MAKEFILE_LIST) | column -tl 2

audio: ## Create MP3 files for text in a CSV file.
	@ go run cmd/audio/main.go

lint: ## Run linters.
	@ golangci-lint run

test: tmp ## Run unit tests.
	@ go test -coverprofile=tmp/cover.out  ./...

tmp: ## Create a tmp directory.
	@ mkdir -p tmp

tui: ## Run the application using a terminal UI.
	@ go run cmd/tui/main.go

web: ## Run the application using a web UI.
	@ . $(HOME)/.nvm/nvm.sh \
		&& nvm install stable \
		&& npm install \
		&& npx tsc
	@ go run cmd/web/main.go
