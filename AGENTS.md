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

## Plans

The user reviews the plan for a feature rather than every line of its code.

### When a plan is needed

Write a plan for any change that involves design decisions, such as a new
endpoint, storage, a page or infrastructure. Changes without any (such as
dependency bumps, typo fixes or small docs changes) skip the plan and go
straight to a PR that's ready for review.

### Steps

1. Open a draft PR whose first commit adds the plan as
   `docs/plans/<short-description>.md` (the branch name without its type),
   and nothing else. The PR description summarises the plan and links to it,
   so that the user can comment on individual lines of the plan.
2. Wait for the user to approve the plan in the PR. Revise it in response to
   their comments, without starting the implementation.
3. Once it's approved, implement it in further commits, then mark the PR as
   ready for review.
4. The user skims the implementation and merges the PR.

### What a plan contains

Keep it concise, with these sections:

- **Goal:** what the change does, and what it leaves for later PRs.
- **Design:** the decisions and why, with alternatives that were rejected.
  Diagrams (Mermaid) are optional.
- **Risks:** what could go wrong, and for each risk, which test or check
  catches it, or that none does.
- **Close review:** which parts need a closer look than a skim (see below),
  or "none".

### Departing from the plan

If the implementation departs from the approved plan, update the plan in the
same PR, and list the departures under "Changes from the plan" in the PR
description, so that the user can see them without comparing versions. Ask
before a departure that changes the design rather than a detail. Once the PR
is merged, the plan is a record of what was decided and isn't updated any
more.

### What still gets a close review

A skim is enough only where a mistake would be caught by a check or easy to
undo. Flag these in the plan's "Close review" section and in the PR
description, pointing at the files or lines concerned:

- authentication and authorization (token verification, the allowlist)
- IAM, service accounts, Workload Identity Federation and secrets
- anything that deletes or migrates data, or could (Firestore writes that
  overwrite, Terraform changes that destroy or replace resources)
- CI/CD workflows, their permissions, and branch rulesets
- new dependencies (as opposed to version bumps)

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
- For front-end changes, check the pages in a browser, and share screenshots
  (light and dark mode, at phone width). Since sign-in needs Firebase, use a
  temporary build that fakes a signed-in user and a mocked API, and don't
  commit it.

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
