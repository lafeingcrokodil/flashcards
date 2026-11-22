# Block merging until the "Before merging" tasks are done

For lafeingcrokodil/flashcards#45.

Markers: 🟡 a judgement call where you might prefer the alternative; 🔴
risky or uncertain.

## Goal

A PR can't be merged while any of its "Before merging" tasks (close reviews,
settings changes, `terraform plan`, approving the "After merging" tasks) is
open, and the plan is approved in the same way. Instead of a checklist in the description and a custom check that
parses it, each task becomes a review thread on the PR, and the `main`
ruleset's built-in "Require conversation resolution before merging" blocks
the merge until every thread is resolved.

This PR changes `AGENTS.md` and the PR template. There's no workflow or
script. Once the plan is approved, I'll update the title and description of
lafeingcrokodil/flashcards#45, which ask for a required check, to match.

diary has nothing like this, so there's nothing to compare with.

## Design

### Why review threads

Resolving threads is enforced by GitHub itself, so there's no parsing to get
wrong, and nothing to test or maintain. Also:

- A close review sits on the lines it's about, so you can resolve it where
  you review it.
- Your own review comments also block the merge until they're resolved. I
  think that's what you'd want, since an unanswered comment is also unfinished
  work.

Rejected (see the [discussion on the earlier version of this
plan](https://github.com/lafeingcrokodil/flashcards/pull/96#discussion_r4657829999)):

- A required check that parses the checklist in the description (the
  earlier version of this plan). It needs a script with its own tests, and
  a parsing bug could let unticked items through without anyone noticing.
- An environment with you as its required reviewer: one approval for the
  whole PR rather than one per task, and every push needs approving again.
- A required approving review: GitHub doesn't let a PR's author approve it,
  and the PRs are opened under your account.
- One issue per task, with a check that they're closed: still a custom
  check, plus an issue for every small task.

### What a task thread looks like

I start a review thread for each task, on the line it's about:

| Task | Where the thread goes |
| --- | --- |
| Approving the plan | The plan's title |
| A close review | The lines to review (the code, or the plan's line for a decision that isn't in the code) |
| A setting to change | The plan's line for the setting, or for a PR without a plan, the line of the change that needs it |
| `terraform plan` | The Terraform file whose resources change |
| Approving the "After merging" tasks | The plan's line for the first task, or for a PR without a plan, the line of the change it checks |

Each thread starts with **Before merging:** and says what to do, briefly,
pointing at the plan's "Close review" section rather than repeating it. It
ends with the usual footer for comments I post.

The thread for approving the plan goes up when I open the draft PR, and
resolving it is how you approve the plan. I start the implementation only
once it's resolved, and if you've left other comments on the plan, I take
the approval to mean that my replies and revisions settle them. Being
resolved, it doesn't hold up the merge later; until then, the PR is a
draft, which can't be merged anyway.

The threads for close reviews go up when I mark the PR as ready for review,
since the code they point at exists only then. Threads for settings can go
up earlier, if a setting is needed before the PR's checks can pass.

### 🟡 Only you resolve threads

`AGENTS.md` says that only you resolve review threads, both the task
threads and the ones you start. Otherwise I could resolve a thread when
I've replied to it, and the merge wouldn't be blocked by anything you
haven't seen. The cost: you resolve threads that I've answered, even when the
answer is just "done".

### 🟡 No "Before merging" section in the description

The PR template's "Before merging" section is removed, so that the tasks are
in one place (the threads) rather than two, which could drift apart. GitHub's
merge box says how many conversations are unresolved and links to them, and
the "Files changed" tab shows them in place.

- Rejected: keeping the section as a list of links to the threads. The links
  wouldn't show which threads are resolved, so it would only repeat the
  threads, which is the kind of duplication you'd rather avoid.

"For you" keeps "After merging" as it is. Approving those tasks is now done
by resolving that thread rather than ticking a box, and that's when I open
their follow-up issues.

### Changes to `AGENTS.md`

- **"PR descriptions":** the "Before merging" checklist is replaced by the
  task threads above, including when they go up and who resolves them. This
  also allows me to post them, since otherwise I only comment on GitHub when
  asked.
- **"What still gets a close review":** each close review gets a task thread,
  instead of an item in the "Before merging" checklist.
- **"Follow-up issues":** the "After merging" tasks are approved by resolving
  their thread.
- **"Plans" → "Steps":** you approve a plan by resolving its approval
  thread.
- **"Workflow":** the example of what you may have edited in a description
  (ticked checklist items) is changed, since there's no checklist any more.

### The ruleset

"Require conversation resolution before merging" is an option of the `main`
ruleset's "Require a pull request before merging" rule, which you turn on in
the settings (Settings → Rules → Rulesets → `main`). That rule is already
on, so this only ticks the option.

## Risks

- **The ruleset option isn't on, or doesn't block the merge:** tested on this
  PR itself. Its task threads stay open until you've turned the option on,
  and the merge box should then say the PR can't be merged until they're
  resolved.
- **I forget to post a task thread:** not caught by any check, the same as
  forgetting a checklist item before. The plan's "Close review" section lists
  what the threads should cover, so a missing one shows when you compare
  them.
- **A thread is resolved without the task being done** (e.g. by mistake):
  the same as ticking a box by mistake. It can be unresolved again.
- **A push makes a thread outdated** (its lines change): outdated threads
  still count as unresolved, so the merge stays blocked.

## Close review

- 🔴 **`AGENTS.md`:** the new rules for task threads, since they govern what
  I do on every PR, and no check tests them.
- **The `main` ruleset:** turning on conversation resolution is done by hand
  in the settings, so nothing in the PR shows it.
- The PR template only needs a skim: a mistake there is easy to fix and
  does little harm.
