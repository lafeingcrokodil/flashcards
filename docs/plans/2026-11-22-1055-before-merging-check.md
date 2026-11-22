# Required check for the "Before merging" checklist

For lafeingcrokodil/flashcards#45.

Markers: 🟡 a judgement call where you might prefer the alternative; 🔴
risky or uncertain.

## Goal

A check that fails while a PR's "Before merging" checklist has unticked
items, required by the `main` ruleset, so that close reviews, settings
changes and approving the "After merging" tasks block the merge until
they're done. Ticking a box edits the PR description, which re-runs the
check.

Left for later: nothing planned. The "After merging" list isn't checked,
since its tasks are tracked in follow-up issues (see `AGENTS.md`).

diary has nothing like this, so there's nothing to compare with.

## Design

### What counts as unticked

The check reads the PR description and fails if the "Before merging"
section has any unticked task list item. In detail:

- **The section** starts at a heading whose text is exactly
  `Before merging` (any level; it's `###` in the template), and ends at the
  next heading of the same or a higher level (`### After merging`,
  `## ...`).
- **HTML comments are removed first.** The template's comment under
  "Before merging" contains example items (`- [ ] <A setting to change>`),
  which GitHub doesn't render, so a PR that keeps the comment mustn't fail
  because of them.
- **An unticked item** is a line like `- [ ] ...`: a list marker (`-`, `*`,
  `+` or `1.`), optionally indented (nested items count too), then `[ ]`.
  `[x]` and `[X]` are ticked.
- **No section, or a section without items, passes.** Dependabot's PRs,
  docs-only PRs and "Nothing beyond skimming and merging." don't need one.

When it fails, the error lists the unticked items, so that the check's
annotation on the PR says what's left without opening the log.

🟡 **No section passes**, rather than failing. Requiring a section would
catch a description that forgot it, but every Dependabot PR would then fail,
and an empty section would pass anyway, so it wouldn't prove anything was
considered.

### Where it runs

- **A new workflow, `.github/workflows/checklist.yml`**, with one job,
  `before-merging`, which is the name the ruleset requires.
- **Triggers:** `pull_request` with the types `opened`, `edited`,
  `reopened` and `synchronize`. `edited` is what re-runs it when a box is
  ticked (it also fires for title and base branch changes, which is
  harmless). `synchronize` (a push) is needed too, since a required check
  must have passed on the PR's latest commit, and a push starts a new
  commit with no result yet.
- **The logic** is in `.github/scripts/check-before-merging.js`, run with
  `actions/github-script` at the same pinned SHA as `vuln.yml` and
  `versions.yml`. It exports the parsing as a separate function from the
  part that reads the PR, so that the parsing can be run locally on its own.
- **Permissions:** `contents: read` (for the checkout) and
  `pull-requests: read`. Nothing is written.

🟡 **It reads the description from the API** (`pulls.get`) rather than from
the event's payload. The payload is the description at the time of the
event, and re-running a job reuses it, so after a re-run the check would
still see an old description. Reading it at run time always uses the
current one. The cost is the `pull-requests: read` permission and one API
call.

- Rejected: `pull_request_target`, which would run the workflow from `main`
  rather than from the PR. It's only needed for write access on PRs from
  forks, which this check doesn't need, and it's the trigger that's easy to
  make unsafe.

🔴 **The description is untrusted input** (anyone who opens a PR writes
it), so it's never interpolated into a `run:` step or the script with
`${{ }}`. The script only reads it as a string from the API's response and
matches it against regular expressions.

### Testing

Like the other scripts in `.github/scripts/`, no committed tests, as agreed
for lafeingcrokodil/flashcards#88. Instead:

- Before marking the PR as ready, I'll run the parsing locally against fake
  descriptions: all ticked, some unticked, nested items, `[X]`, the
  template's comment left in, no section, an empty section, and an
  unticked item under "After merging" (which must pass).
- On this PR itself: the check runs on its own description, so it should
  fail while this PR's "Before merging" items are unticked, and pass once
  you tick them, without a push. That's the end-to-end test of the `edited`
  trigger and of the ruleset.

🟡 A committed test with Node's built-in test runner (`node --test`, no
dependencies) would catch a regression in the parsing, which matters more
here than for the other scripts: a bug that lets unticked items through
fails silently. It would need a Node setup step in CI, though, and it'd be
the first JavaScript test, so I've kept to the agreed approach. Say if
you'd rather have it.

### The ruleset

`before-merging` is added to the `main` ruleset's required checks by hand,
like `test`, `lint` and `govulncheck`, once it has run on this PR (GitHub
only offers checks that have run before).

### `AGENTS.md`

A line under "PR descriptions": the `before-merging` check fails while the
"Before merging" checklist has unticked items, and only you tick them.
That's what makes the check mean something: if I ticked items myself when
updating a description, it would pass without your review.

## Risks

- **Unticked items are missed, so the check passes when it shouldn't**
  (e.g. a list marker or heading format the regular expressions don't
  match): covered by the local run against fake descriptions before
  merging, not by a committed test (see "Testing"). 🔴 A regression later
  would pass silently.
- **The check fails when it shouldn't** (e.g. on the template's comment):
  covered by the same local run. A false failure blocks the merge, so it's
  noticed at once, and you can bypass the ruleset meanwhile.
- **Ticking a box doesn't re-run the check:** tested on this PR itself (see
  "Testing").
- **An old run's result counts** (e.g. two quick edits, and the first
  run finishes last): GitHub uses the most recently created check run of
  that name on the commit, which is the run of the later edit, and that run
  reads the description when it runs, so it sees both edits.
- **The check can't run while the PR has a merge conflict** (GitHub doesn't
  run `pull_request` workflows then): harmless, since a conflicting PR
  can't be merged anyway, and resolving it pushes a commit, which re-runs
  the check.
- **The ruleset doesn't require the check:** not caught by any check, since
  the ruleset is outside the repository. Flagged for close review.

## Close review

- **`.github/workflows/checklist.yml`:** its triggers and permissions, and
  that the description isn't interpolated anywhere.
- 🔴 **`.github/scripts/check-before-merging.js`:** the parsing rules, since
  a mistake that lets unticked items through isn't caught by any committed
  check.
- **The `main` ruleset:** adding `before-merging` as a required check is
  done by hand in the settings.
