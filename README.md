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
- for the front end, and for `make web`, `make run` and `make build`, which
  build it: the requirements in [`web/README.md`](web/README.md)

Run `make help` (or just `make`) to list the available targets.

The Go binary embeds the built front end (`web/build/`) and serves it, so the
whole app is a single binary. `make run` builds the front end and starts the
server on http://localhost:8080, and `make build` builds a binary that
includes it.

`make test`, `make cover`, `make vet`, `make lint` and `make vuln` don't need
the front end: if it hasn't been built, they create a placeholder page so that
the Go code compiles.

### Working on the front end

For instant reloading while you edit the front end, run the Go server and
Vite's dev server side by side:

```bash
make run
```

```bash
cd web && pnpm dev
```

Then open http://localhost:5173. Vite forwards requests under `/api` to the Go
server.

### Automated checks

| Workflow | When | What it does |
| --- | --- | --- |
| [`ci.yml`](.github/workflows/ci.yml) | every pull request and push to `main` | `make vet`, `make cover` and golangci-lint, and the front end's `pnpm check`, `pnpm lint` and `pnpm build` |
| [`vuln.yml`](.github/workflows/vuln.yml) | every pull request and push to `main`, and daily | `make vuln` |
| [`versions.yml`](.github/workflows/versions.yml) | monthly | checks whether Go or golangci-lint is behind the newest release that's at least 14 days old |

When a daily or monthly run finds a problem, it opens an issue with the
`automated` label, which is closed automatically once the check passes again.
