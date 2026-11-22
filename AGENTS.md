# Working on this repo

## Workflow

- Work incrementally: each PR should be a small, self-contained change that
  is easy to review.
- Keep `docs/roadmap.md` up to date as milestones progress.
- Plan features before implementing them, as described below.
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

## Reference repository

[lafeingcrokodil/diary](https://github.com/lafeingcrokodil/diary) is a
similar app with the same stack. Treat it as a reference: before designing
something that it already has (such as the Makefile, the lint configuration,
CI, the Dockerfile or the Terraform layout), look at how it does it.

Departing from it is fine, and desired where its approach has a problem, but
flag each departure and give the reason: in the plan's "Design" section, or
in the PR description for a change without a plan.

## Plans

The user reviews the plan for a feature rather than every line of its code.

### When a plan is needed

Write a plan for any change that involves design decisions, such as a new
endpoint, storage, a page or infrastructure. Changes without any (such as
dependency bumps, typo fixes or small docs changes) skip the plan and go
straight to a PR that's ready for review.

### Steps

1. Open a draft PR whose first commit adds the plan as
   `docs/plans/<time>-<short-description>.md`, and nothing else. `<time>` is
   when the plan is written, in UTC as `YYYY-MM-DD-HHMM`, so that plans sort
   chronologically, and `<short-description>` is the branch name without its
   type (`feat/healthz-endpoint` → `2026-11-21-1405-healthz-endpoint.md`).
   The PR description links to the plan, so that the user comments on
   individual lines of the plan itself (see "PR descriptions" below).
2. Wait for the user to approve the plan in the PR. Revise it in response to
   their comments, without starting the implementation.
3. Once it's approved, implement it in further commits, then mark the PR as
   ready for review.
4. The user skims the implementation and merges the PR.

### What a plan contains

Keep it concise, with these sections:

- **Goal:** what the change does, and what it leaves for later PRs.
- **Design:** the decisions and why, with alternatives that were rejected.
- **Risks:** what could go wrong, and for each risk, which test or check
  catches it, or that none does.
- **Close review:** which parts need a closer look than a skim (see below),
  or "none".

Add whatever else helps to review the plan, such as diagrams (Mermaid) or UI
mockups. Put supplementary files, such as images, in a directory named like
the plan without `.md` (`docs/plans/2026-11-21-1405-healthz-endpoint/`).

### Departing from the plan

If the implementation departs from the approved plan, update the plan in the
same PR, and list the departures under "Changes from the plan" in the PR
description, so that the user can see them without comparing versions. Ask
before a departure that changes the design rather than a detail. Once the PR
is merged, the plan is a record of what was decided and isn't updated any
more.

### What still gets a close review

A skim is enough only where a mistake would be caught by a check, or would be
easy to undo and do little harm. Everything else needs a close review.

To fill in the "Close review" section, go through each part of the change
(each file, decision or behaviour) and ask:

1. If this were wrong, would a test, linter or other check catch it?
2. If not, would it be easy to undo, and do little harm in the meantime?

Flag every part where both answers are no in the plan's "Close review"
section, pointing at the files or lines concerned.
Write "None" only if every part passes, and say briefly why the parts that
come closest do.

Some things almost always need a close review, but they're examples, not a
checklist: a change can need one without being on this list.

- authentication and authorization (token verification, the allowlist)
- IAM, service accounts, Workload Identity Federation and secrets
- anything that deletes or migrates data, or could (Firestore writes that
  overwrite, Terraform changes that destroy or replace resources)
- CI/CD workflows, their permissions, and branch rulesets
- new dependencies (as opposed to version bumps)

### PR descriptions

For a PR with a plan, the description only adds what the plan doesn't say,
so that the user doesn't read the same thing twice and the two can't drift
apart:

- a link to the plan, and to its "Close review" section once the PR is ready
  for review
- "Changes from the plan" (see above), or that there are none
- the checks that ran and what they showed (see below)
- anything the user has to do, such as a setting to change before merging
- anything that came up during the implementation that's worth knowing but
  doesn't belong in the plan

Don't summarise the plan. A PR without a plan has its description say what
the change does and why, as usual.

### Checks before a skim

Before marking a PR as ready for review:

- All CI checks pass. Until CI exists, run the equivalent checks locally and
  say in the PR description which ones ran.
- The tests cover the behaviour described in the plan, including its error
  cases, and each risk in the plan is either covered by a check or flagged
  for close review.
- For front-end changes, the PR has screenshots (see below).
- For Terraform changes, the PR says which resources should change, so that
  the user can compare that with `terraform plan`.

## Code comments

Keep comments within the scope of the code they describe. A comment should
explain what the code does and why, without relying on details of other parts
of the system (such as where it's deployed or who calls it), so that it doesn't
become stale when those other parts change.

## Linting and testing

- Linters run with all rules enabled. Only disable a rule with a comment that
  explains why.
- In Go tests, use testify for assertions: `require` when the test can't
  usefully go on after a failure (such as an unexpected error), and `assert`
  otherwise.
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

## Branch names

Name branches after their purpose, in the form `<type>/<short-description>`,
using kebab-case. For example:

- `feat/healthz-endpoint`
- `ci/github-actions`
- `fix/review-scheduling`
- `docs/roadmap`

Use this convention instead of generated branch names (such as
`claude/<random-words>`). Pick the name when starting the work, and use a new
branch for each PR.

## Dependencies

Only use dependency versions that were published at least 14 days ago (Go
modules, GitHub Actions, npm packages, Terraform providers and container
images).
