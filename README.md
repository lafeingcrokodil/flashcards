# flashcards

A web app for learning languages, with flashcards and reading practice.

The app is being rebuilt from scratch. Until then,
[`prototype/`](prototype/README.md) shows the features it's meant to have.
The Terraform configuration for its infrastructure lives in
[`infra/`](infra/README.md).

## Development

Requires:

- the Go version specified in `go.mod`, or later
- [golangci-lint](https://golangci-lint.run/docs/welcome/install/) v2.15.0
  (for `make lint`)

```bash
make run    # start the server on http://localhost:8080 (or the port in PORT)
make build  # build a server binary (bin/flashcards)
make test   # run the tests
make cover  # run the tests and check the coverage of each package in internal/ (see Makefile)
make vet    # run go vet
make lint   # run golangci-lint (see .golangci.yml)
```

CI (`.github/workflows/ci.yml`) runs `make vet`, `make cover` and
golangci-lint on every pull request and every push to `main`.
