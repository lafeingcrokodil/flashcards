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
  5 seconds for in-flight requests (`Server.Shutdown`). That's shorter than
  the 10 seconds Cloud Run waits after `SIGTERM` before killing the
  container, so that the server exits cleanly rather than being killed.
- So that shutdown can be tested, `main` only calls
  `run(ctx, getenv, stderr) error`, which sets up signal handling
  (`signal.NotifyContext`) and the server. Serving is a separate function,
  `serve(ctx, listener, handler, shutdownTimeout) error`, which returns once
  `ctx` is cancelled and in-flight requests have finished (or the timeout has
  passed).
- Logs with `log/slog`'s JSON handler to stderr, since structured logs are
  easier to filter. Only startup, shutdown and errors are logged for now; no
  request logging yet.

### Linting

`.golangci.yml` enables all linters (`default: all`), per `AGENTS.md`, and
each disabled one has a comment explaining why:

- `exhaustruct` and its replacement `exhaustruct_v5`: they require every
  field of a struct literal to be set, including optional ones whose zero
  values are the intended defaults (such as most fields of `http.Server`).
- Deprecated linters whose replacements stay enabled: `gomodguard`
  (→ `gomodguard_v2`) and `wsl` (→ `wsl_v5`).

`depguard` stays enabled, with an allowlist of the standard library and this
module, so that a new dependency fails the linter until it's added to the
list deliberately.

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
- **Graceful shutdown doesn't work, or breaks later:** covered by tests, so
  that a regression fails CI once it exists:
  - `serve` with a handler that blocks until the test releases it: cancelling
    `ctx` mid-request still lets the request complete with `200`, new
    connections are refused, and `serve` returns `nil`.
  - `serve` with a request that outlasts the shutdown timeout: `serve`
    returns an error instead of hanging.
  - `run` on port `0`: the test sends `SIGTERM` to its own process (which
    `signal.NotifyContext` catches, so the test binary isn't killed) and
    checks that `run` returns `nil` within a few seconds. This covers the
    signal wiring, which the `serve` tests don't.
- **Lint config hides real problems:** every disabled linter has a comment
  explaining why, and the list is short enough to skim.
- **A dependency is added without a close review:** `depguard` fails on any
  import outside the standard library and this module.
- **The local golangci-lint can't lint Go 1.27:** the one in this environment
  is v2.5.0, built with Go 1.25, so v2.15.0 (published 2026-10-16) was
  installed with Go 1.27.4 instead.

## Close review

There's no auth, data, infrastructure, CI or new dependency in this PR, but
these parts aren't covered by a check and are harder to change later:

- **The module path** (`go.mod`, and the imports that use it): changing it
  later means rewriting every import, and once a `v3.x` tag is pushed, the
  path is effectively fixed. Worth confirming `/v3` now.
- **The disabled linters and the `depguard` allowlist** (`.golangci.yml`): no
  check notices if a linter is disabled, or a package allowed, without good
  reason, and code written while a linter is off can pile up findings that
  make it harder to turn back on.

The rest (routes, timeouts, Makefile targets) is covered by tests or easy to
change.
