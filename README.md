# flashcards

A web app for learning languages, with flashcards and reading practice.

The app is being rebuilt from scratch. Until then,
[`prototype/`](prototype/README.md) shows the features it's meant to have.
The Terraform configuration for its infrastructure lives in
[`infra/`](infra/README.md).

[`docs/architecture.md`](docs/architecture.md) describes how the app is built,
and the [milestones](https://github.com/lafeingcrokodil/flashcards/milestones)
track the planned work.

## Development

Requires:

- the Go version specified in `go.mod`, or later
- [golangci-lint](https://golangci-lint.run/docs/welcome/install/) v2.15.0
  (for `make lint`)

Run `make help` (or just `make`) to list the available targets.

CI (`.github/workflows/ci.yml`) runs `make vet`, `make cover` and
golangci-lint on every pull request and every push to `main`.
`.github/workflows/vuln.yml` runs `make vuln` on the same events and daily.
When a daily run fails, it opens an issue with the `automated` label, which is
closed automatically once the check passes again.
`.github/workflows/versions.yml` checks monthly whether Go or golangci-lint is
behind the newest release that's at least 14 days old, and reports it the
same way.
