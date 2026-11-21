# GitHub Actions CI

## Goal

Run the Go checks on every pull request and every push to `main`: `go vet`,
the tests with a minimum coverage per package in `internal/`, and
golangci-lint. The `main` ruleset then requires these jobs to pass before a
PR can be merged.

Left for later PRs: Dependabot (the next item), front-end jobs (with the
SvelteKit scaffold), building the image (with the Dockerfile), Terraform
checks and deployment (M1).

## Design

### Workflow

`.github/workflows/ci.yml`, modelled on the diary app's:

```yaml
on:
  push:
    branches: [main]
  pull_request:

permissions:
  contents: read

jobs:
  test:   # checkout, setup-go (version from go.mod), make vet, make cover
  lint:   # checkout, setup-go, golangci-lint-action with version v2.15.0
```

- **Triggers:** pull requests, and pushes to `main` so that `main` itself is
  checked after a merge. Pushes to other branches aren't checked on their
  own, since they're checked through their PR.
- **Permissions:** `contents: read` for the whole workflow. Nothing writes to
  the repository or needs other tokens.
- **Two jobs, `test` and `lint`**, as in diary, so that a lint failure and a
  test failure show separately, and they run in parallel. `go vet` runs in
  `test` (it's quick, and it needs no extra setup).
- **The jobs call Make targets** (`make vet`, `make cover`), so that CI runs
  the same commands as a local run. golangci-lint runs through its action,
  which caches its results between runs.
  - Rejected: `make lint` in CI, which would need golangci-lint to be
    installed in a separate step, without the action's cache.
- **golangci-lint v2.15.0**, the version used locally so far, pinned in the
  action's `version` input. `AGENTS.md` already points to the `lint` job for
  the version to install locally, and the README will mention it too.

### Pinning actions

Each action is pinned to a full commit SHA, with the version in a comment
(Dependabot keeps both up to date):

| Action | Version | Released (tagged commit) |
| --- | --- | --- |
| `actions/checkout` | v7.0.1 | 2026-07-17 |
| `actions/setup-go` | v7.0.0 | 2026-07-15 |
| `golangci/golangci-lint-action` | v9.3.0 | 2026-06-29 |

Departure from diary, which uses major-version tags (`@v7`). A major tag
moves to each new release as soon as it's published, which breaks the 14-day
rule: `actions/setup-go@v7` currently points to v7.1.0, published
2026-11-16. A tag can also be moved to different code by whoever controls
the action's repository, while a SHA can't.

`actions/setup-go` v7.1.0 is skipped for now, since it's less than 14 days
old.

### Coverage

A `make cover` target runs the tests with a coverage profile and fails if any
package in `internal/` has less than 80% of its statements covered. It uses
the same `awk` script as diary's Makefile, which reads the profile and adds
up the statements per package. `cmd/` isn't checked: it only wires things
together, and its tests are about behaviour (shutdown) rather than coverage.
`make test` stays as it is, for a quicker local run.

Departure from diary: there's no list of test helper packages to skip yet,
since there are none. The list can be added with the first one.

`coverage.out` is added to `.gitignore`.

### Ruleset

The `main` ruleset should require the status checks `test` and `lint` (the
job names). I can't change rulesets from this environment, so you'd add them
in the repository's settings (Settings → Rules → Rulesets → `main` →
"Require status checks to pass") once the jobs have run on this PR, since
GitHub only offers checks that have run before. I'll say in the PR
description exactly what to add.

### README

A short "Development" section in the README, like diary's: what's needed
(Go from `go.mod`, golangci-lint v2.15.0) and what each Make target does.

## Risks

- **A check is skipped or fails silently, so CI passes without running it:**
  the PR's own CI run shows each step's output; I'll check that `make vet`,
  `make cover` and golangci-lint ran and passed. To check that failures fail
  the run, I'll push a temporary commit with a lint error and one that drops
  coverage below 80%, see both jobs fail, and then revert them.
- **The coverage script miscounts:** checked by hand against
  `go tool cover -func` for `internal/server` (100% now). A later drop in
  coverage is what the target is for, so no separate test of the script.
- **The workflow gets more permissions than it needs:** set explicitly to
  `contents: read`, and visible in the YAML.
- **An action is pinned to a version less than 14 days old, or the wrong
  SHA:** the table above lists the versions and dates, and the SHAs can be
  compared with each action's tags. Dependabot's cooldown (next item) keeps
  later updates to the rule.
- **The ruleset doesn't require the checks, so a failing PR can still be
  merged:** not caught by any check, since the ruleset is outside the
  repository. Flagged for close review.

## Close review

- **`.github/workflows/ci.yml`:** its triggers and permissions, and that the
  actions' SHAs match the versions in their comments.
- **The `main` ruleset:** the change is made by hand in the settings, so
  nothing in the PR shows it. Worth checking afterwards that a PR with a
  failing check can't be merged.
