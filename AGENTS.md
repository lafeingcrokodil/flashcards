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
  Before each update, read the current version and build on it, since the
  user may have edited it since (e.g. to add a note), and an update replaces
  the whole description. The same goes for issue descriptions.
- Subscribe to the activity (comments, reviews, CI results) of each PR you
  open, straight after opening it, without asking first. That way, review
  comments and CI failures get a response without the user having to point
  them out. If the environment can't subscribe, say so in chat.
- Don't merge PRs, even once it's agreed that one should be merged: the
  user merges them.
- Before every force-push, ask the user in chat, saying exactly what will be
  pushed where.
- The user runs `terraform plan` and `terraform apply`.
- Chat is for what isn't on GitHub. Before each message, check every
  sentence: if it says something a PR, issue, plan, comment or thread
  already says, or describes something done there (opening, updating,
  replying, resolving, pushing), cut it, and link instead. That includes
  pointing out parts of a plan: its 🟡 and 🔴 markers already do that.
  What's left is what needs the user and isn't on GitHub, such as a
  question that blocks the work, or something found outside the PR's scope.
  If nothing's left, the message is just the link, or nothing at all when
  a notification changed nothing.
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
   closes (see "PR descriptions"). A task thread on the plan's title asks
   the user to approve it (see "Task threads").
2. Wait for the user to approve the plan by resolving that thread. Revise it
   in response to their comments, without starting the implementation.
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
section, pointing at the files or lines concerned, and start a task thread
for each one (see "Task threads").
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
- Each close review in the plan has its task thread.

## PR titles

Keep PR titles to at most 65 characters, and say what the change does rather
than listing everything in it. PRs are squash-merged, so the title becomes the
commit's subject line, with ` (#N)` appended, and 65 characters keep that
within Git's usual 72.

## PR descriptions

Follow `.github/pull_request_template.md`. In particular:

- For a PR with a plan, the description only adds what the plan doesn't say,
  so that the user doesn't read the same thing twice and the two can't drift
  apart. Don't summarise the plan: link to it.
- Tasks for the user before merging are task threads (see below), not part
  of the description, so that they're in one place.
- The "Follow-ups" section lists the tasks after merging, for the user or
  for an agent, as a plain list rather than a checklist, since each task is
  tracked in its own issue (see "Follow-up issues"): checks that can only be
  done once the change is on `main` (such as whether Dependabot picked up
  its configuration), and `terraform apply` for Terraform changes.

### Task threads

The `main` ruleset blocks merging while a PR has unresolved review threads.
So each task for the user before merging is a review thread on the PR, which
the user resolves once it's done. Start one, on the line it's about, for:

| Task | Where the thread goes |
| --- | --- |
| Approving the plan | The plan's title |
| A close review | The lines to review (the code, or the plan's line for a decision that isn't in the code) |
| A setting to change before the PR can work | The plan's line for the setting, or for a PR without a plan, the line of the change that needs it |
| `terraform plan`, to compare with the resources the PR says should change | The Terraform file whose resources change |
| Adding a new CI job to the `main` ruleset's required checks, once it has run on the PR | The job's name in the workflow |
| Approving the follow-ups | The plan's line for the first task, or for a PR without a plan, the line of the change it checks |
| Anything else that has to be done first | The line it's about |

- Each thread starts with **Before merging:** and says briefly what to do,
  pointing at the plan (e.g. its "Close review" section) rather than
  repeating it.
- The plan's thread goes up with the draft PR. The other threads go up when
  the PR is marked as ready for review, since the code they point at exists
  only then, or earlier for a setting that the PR's checks need.
- Only the user resolves threads: task threads, and the ones the user
  starts, even once they've been answered. Otherwise a thread could be
  resolved without the user having seen it.

### Follow-up issues

A task in a PR description is easy to lose once the PR is merged, so each
follow-up gets a GitHub issue, and the issue is where it's marked as done:

- Until the user has approved the follow-ups (by resolving that task
  thread), list them without issues, so that the user can change them
  first. Once approved, open an issue for each, with the `follow-up` label
  and without an assignee, since either the user or an agent may do it.
  Its description says what to do and links to the PR. Then replace each
  task in the PR description with just the issue's reference
  (`lafeingcrokodil/flashcards#N`), since GitHub shows a reference on its
  own in a list as the issue's title.
- If the PR is closed without being merged, close its follow-up issues too.
- Anything else that's left to do, such as a limitation found during a PR
  that's out of its scope, also gets an issue: in a milestone if it's
  planned work, or without a milestone if it isn't planned yet (an idea). The
  `follow-up` label is only for tasks after a PR.
- Issues with the `automated` label are opened and closed by scheduled
  checks (with `.github/scripts/report-issue.js`), one per check: leave them
  to the check, which closes its issue once it passes again.

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

- Before every push, run all of CI's checks locally (for the Go code, that
  includes `make lint`), so that problems show up before CI runs rather than
  after. If a tool that a check needs is missing or too old, install it
  rather than leaving the check to CI. If it can't be installed, say so in
  chat, and don't push until the user has agreed how to proceed.
- `golangci-lint` must be built with a Go version at least as new as the one
  in `go.mod`. If the installed one is older, install the version that CI
  uses (the `lint` job in `.github/workflows/ci.yml`) with the toolchain
  that `go.mod` asks for. `go install` with a version ignores `go.mod`, so set
  the toolchain explicitly, from the repository's root:
  `GOTOOLCHAIN=$(go env GOVERSION) go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@<version>`.

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
