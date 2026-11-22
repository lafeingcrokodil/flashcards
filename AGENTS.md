# Working on this repo

## Workflow

- Work incrementally: each PR should be a small, self-contained change that
  is easy to review.
- Track planned work as issues in the repository's
  [milestones](https://github.com/lafeingcrokodil/flashcards/milestones),
  one issue per PR-sized item. A PR that finishes an issue says
  `Closes lafeingcrokodil/flashcards#N` in its summary, so that merging it
  closes the issue. New planned work gets a new issue in the right
  milestone.
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
   individual lines of the plan itself, and names the issue that the PR
   closes (see "PR descriptions").
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

Mark the points that most need the user's attention, so that they stand out
when skimming the plan:

- 🟡 a judgement call: there's a reasonable alternative, and the user might
  prefer it.
- 🔴 risky or uncertain: not covered by a check and hard to undo (so it's in
  "Close review" too), or based on something that couldn't be verified.

Use them sparingly: a plan where everything is marked has nothing marked.

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
section, pointing at the files or lines concerned, and add each one to the
PR description's "Before merging" checklist (see "PR descriptions").
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

## PR descriptions

Follow `.github/pull_request_template.md`. In particular:

- For a PR with a plan, the description only adds what the plan doesn't say,
  so that the user doesn't read the same thing twice and the two can't drift
  apart. Don't summarise the plan: link to it.
- Everything that needs the user's action goes in the "For you" section,
  split by when it's due:
  - **Before merging**, as a checklist: each close review (a short pointer
    to the files and the plan's "Close review" section, not a copy of it),
    settings to change before the PR can work, `terraform plan` for
    Terraform changes (to compare with the resources the PR says should
    change), and anything else that has to be done first. If there are
    "After merging" tasks, the last item is to approve them.
  - **After merging**, as a plain list rather than a checklist, since each
    task is tracked in its own issue (see below): checks that can only be
    done once the change is on `main` (such as whether Dependabot picked up
    its configuration), and `terraform apply` for Terraform changes.

### Follow-up issues

A task in a PR description is easy to lose once the PR is merged, so each
"After merging" task gets a GitHub issue, and the issue is where it's marked
as done:

- Until the user has approved the "After merging" tasks (by ticking that
  item in "Before merging"), list them without issues, so that the user can
  change them first. Once approved, open an issue for each, with the
  `follow-up` label, assigned to the user. Its description says what to do
  and links to the PR. Then replace each task in the PR description with
  just the issue's reference (`lafeingcrokodil/flashcards#N`), since GitHub
  shows a reference on its own in a list as the issue's title.
- If the PR is closed without being merged, close its follow-up issues too.
- Anything else that's left to do, such as a limitation found during a PR
  that's out of its scope, also gets an issue: in a milestone if it's
  planned work, or with the `idea` label if it isn't planned yet. The
  `follow-up` label is only for the user's tasks after a PR.

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
- `docs/architecture`

Use this convention instead of generated branch names (such as
`claude/<random-words>`). Pick the name when starting the work, and use a new
branch for each PR.

## Dependencies

Only use dependency versions that were published at least 14 days ago (Go
modules, GitHub Actions, npm packages, Terraform providers and container
images).
