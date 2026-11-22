# flashcards

A web app for learning languages, with flashcards and reading practice.

The app is being rebuilt from scratch. Until then,
[`prototype/`](prototype/README.md) shows the features it's meant to have.
The Go back end lives in the repository root, the SvelteKit front end in
[`web/`](web/README.md), and the Terraform configuration for its
infrastructure in [`infra/`](infra/README.md).

[`docs/architecture.md`](docs/architecture.md) describes how the app is built,
and the [milestones](https://github.com/lafeingcrokodil/flashcards/milestones)
track the planned work.

## Development

Requires:

- the Go version specified in `go.mod`, or later
- [golangci-lint](https://golangci-lint.run/docs/welcome/install/) v2.15.0
  (for `make lint`)
- for the front end, the requirements in [`web/README.md`](web/README.md)

Run `make help` (or just `make`) to list the available targets.

### Automated checks

| Workflow | When | What it does |
| --- | --- | --- |
| [`ci.yml`](.github/workflows/ci.yml) | every pull request and push to `main` | `make vet`, `make cover` and golangci-lint, and the front end's `pnpm check`, `pnpm lint` and `pnpm build` |
| [`vuln.yml`](.github/workflows/vuln.yml) | every pull request and push to `main`, and daily | `make vuln` |
| [`versions.yml`](.github/workflows/versions.yml) | monthly | checks whether Go or golangci-lint is behind the newest release that's at least 14 days old |

When a daily or monthly run finds a problem, it opens an issue with the
`automated` label, which is closed automatically once the check passes again.
