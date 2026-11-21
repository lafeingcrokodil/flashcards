# Go module and `/healthz` endpoint

## Goal

Start the Go back end for M0's first item: a Go module, an HTTP server with a
`GET /healthz` endpoint, its tests, a golangci-lint configuration and a
Makefile to run them.

Left for later PRs: CI (the next item), Dependabot, the front end and serving
it, the Dockerfile, and anything that needs storage or auth.

## Design

### Module

- Path `github.com/lafeingcrokodil/flashcards/v3`. The old app was
  `.../flashcards/v2` and was released up to `v2.3.2`; the rebuild is a new
  major version, so a release of it would be `v3.0.0`, which Go requires to
  live under `/v3`.
  - Rejected: keeping `/v2` (a later `v2.x` tag would suggest the rebuild is
    compatible with the old app) and dropping the suffix (tags `v2` and above
    can't be used without it).
- `go 1.27.4`, the latest Go release (published 2026-11-02).
- No dependencies: the standard library covers routing (`http.ServeMux` with
  method patterns), logging (`log/slog`) and testing. Tests use the standard
  `testing` package rather than testify, so that this PR adds no dependency.

### Layout

```
cmd/flashcards/main.go     # reads config, starts the server, handles signals
internal/server/server.go  # New() http.Handler with the routes
internal/server/server_test.go
.golangci.yml
Makefile
```

`internal/server` holds the routes so that they can be tested with
`httptest` without starting a real server. Later areas (packs, progress, ...)
get their own packages under `internal/`.

### `/healthz`

- `GET /healthz` returns `200 OK` with `Content-Type: text/plain` and the
  body `ok`. Health checks only look at the status, so a JSON body would add
  nothing.
- Other methods get `405 Method Not Allowed` and unknown paths `404`, both from
  `ServeMux` itself.
- It doesn't check dependencies (there are none yet). If a readiness check for
  Firestore is needed later, it can be a separate endpoint, so that a
  Firestore outage doesn't make the container look dead.

### Server

- Listens on the port in `PORT` (the convention Cloud Run uses), defaulting
  to `8080`. An invalid value is an error at startup, not a silent fallback.
- `http.Server` with `ReadHeaderTimeout` set (gosec flags servers without it).
- On `SIGINT` or `SIGTERM`, it stops accepting connections and waits up to
  10 seconds for in-flight requests (`Server.Shutdown`).
- Logs with `log/slog`'s JSON handler to stderr, since structured logs are
  easier to filter. Only startup, shutdown and errors are logged for now; no
  request logging yet.

### Linting

`.golangci.yml` enables all linters (`default: all`), per `AGENTS.md`. Any
that are disabled get a comment explaining why; the expected ones are
`depguard` (needs an allowlist of packages to be useful),
`exhaustruct` (would require every field of e.g. `http.Server` to be set) and
deprecated linters that have a replacement which is also enabled (`wsl` →
`wsl_v5`). Others are only disabled if they turn out to conflict.

### Makefile

| Target | Runs |
| --- | --- |
| `help` (default) | lists the targets |
| `test` | `go test -race ./...` |
| `lint` | `go vet ./...` and `golangci-lint run` |
| `run` | `go run ./cmd/flashcards` |
| `build` | `go build -o bin/flashcards ./cmd/flashcards` |

`bin/` is added to `.gitignore`.

## Risks

- **`/healthz` or routing behaves differently than described:** covered by
  table-driven `httptest` tests (status, content type and body for `GET`;
  `405` for `POST`; `404` for an unknown path).
- **Invalid `PORT` handled wrongly:** the port parsing is a small function
  with tests for unset, valid and invalid values.
- **Graceful shutdown doesn't work:** not covered by a test (it'd need to send
  a signal to a real process). Checked by hand: run the server, send
  `SIGTERM`, see it log the shutdown and exit with status 0.
- **Lint config hides real problems:** every disabled linter has a comment
  explaining why, and the list is short enough to skim.
- **The local golangci-lint can't lint Go 1.27:** the one in this environment
  is v2.5.0, built with Go 1.25. I'll install a release that supports
  Go 1.27 and is at least 14 days old; if that's blocked, I'll stop and ask.

## Close review

None. There's no auth, data, infrastructure, CI or new dependency in this PR.
