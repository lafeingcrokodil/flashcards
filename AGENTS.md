# Working on this repo

## Workflow

- Work incrementally: each PR should be a small, self-contained change that
  is easy to review.
- Keep `docs/roadmap.md` up to date as milestones progress.
- Keep PR titles and descriptions up to date: when a PR changes after it's
  opened (e.g. in response to review feedback), update them to match.
- Don't merge PRs, even once it's agreed that one should be merged: the
  user merges them.
- Before every force-push, ask the user in chat, saying exactly what will be
  pushed where.
- The user runs `terraform plan` and `terraform apply`.
- After opening a PR, only mention in chat what isn't already in its
  description.
- If something is blocked (e.g. a domain by the network policy, or an action
  by a permission check), stop and ask the user rather than working around
  it. Flag limitations of the environment clearly.

## Code comments

Keep comments within the scope of the code they describe. A comment should
explain what the code does and why, without relying on details of other parts
of the system (such as where it's deployed or who calls it), so that it doesn't
become stale when those other parts change.

## Linting and testing

- Linters run with all rules enabled. Only disable a rule with a comment that
  explains why.
- For front-end changes, check the pages in a browser, and share screenshots
  (light and dark mode, at phone width). Since sign-in needs Firebase, use a
  temporary build that fakes a signed-in user and a mocked API, and don't
  commit it.

## Development environment

- `golangci-lint` must be built with a Go version at least as new as the one
  in `go.mod`. If the installed one is older, install the version that CI
  uses (the `lint` job in `.github/workflows/ci.yml`) with the current
  toolchain, e.g.
  `go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@<version>`.
- The Firestore tests need the emulator (`make emulator`) and fail without
  it: run them with `FIRESTORE_EMULATOR_HOST=localhost:8085 make cover`.

## Branch names

Name branches after their purpose, in the form `<type>/<short-description>`,
using kebab-case. For example:

- `feat/healthz-endpoint`
- `ci/github-actions`
- `fix/diary-date-timezone`
- `docs/roadmap`

Use this convention instead of generated branch names (such as
`claude/<random-words>`). Pick the name when starting the work, and use a new
branch for each PR.

## Dependencies

Only use dependency versions that were published at least 14 days ago (Go
modules, GitHub Actions, npm packages, Terraform providers and container
images). This is enforced where the tooling supports it: pnpm's
`minimumReleaseAge` and Dependabot's cooldown.

The one exception is the distroless runtime image in the `Dockerfile`. Its
images have no publication date, so Dependabot's cooldown can't apply, and its
tag is republished often enough that the newest digest would rarely be 14 days
old. Its digest updates can be merged as they come, since the image only
contains CA certificates, time zone data and user account files.
