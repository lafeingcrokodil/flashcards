# Roadmap as GitHub milestones and issues

Markers: 🟡 a judgement call where you might prefer the alternative; 🔴
risky or uncertain.

## Goal

Track all planned work in one place: GitHub issues, grouped into milestones,
instead of a checklist in `docs/roadmap.md` plus follow-up issues. A PR that
finishes an item closes its issue, so the milestone's progress updates
without anyone editing a file.

The parts of `docs/roadmap.md` that describe the app rather than tasks (the
introduction and the architecture) stay in the repository, in
`docs/architecture.md`.

Left for later: a GitHub Project board, if ordering or a board view turns
out to be needed (see "Design").

## Design

### Milestones

One GitHub milestone per roadmap milestone, with the same name and a short
description. M7's is the roadmap's own; the others are new, written from the
roadmap's items and the prototype:

| Milestone | Description |
| --- | --- |
| M0 – Walking skeleton | The smallest app that builds, tests and runs end to end: a Go server and a SvelteKit front end in one container, with CI and dependency updates. No features yet. |
| M1 – Deployed to GCP | The app runs on Cloud Run in the existing GCP project, deployed from `main`, with its infrastructure managed by Terraform and Firestore set up for the data to come. |
| M2 – Authentication | Only allowed users can use the app: they sign in with Google through Firebase Authentication, and the server checks their ID token against an allowlist. |
| M3 – Packs | Packs of words, each word with its translation, an example sentence and a note, which can be created and edited in the app. They're what Cards, Type and Read build on. |
| M4 – Cards | Flashcard review sessions with Leitner scheduling: due words first, then new ones, with progress stored per word and shown per pack. |
| M5 – Type | Typed recall in the other direction, sharing Cards' scheduling, with alternative answers, a hint and an "I was right" override. |
| M6 – Read | Short passages to read, translate line by line or answer questions on, with a glossary for tapping words and saving them to a pack. |
| M7 – Feedback, lookups and speech | Features that the prototype implements with Claude. They might use an AI model in the app too, but don't have to. |

No due dates: the app is built when there's time, not to a schedule.

None of my GitHub tools can create milestones, so you'd create these eight
in the UI (Issues → Milestones → New milestone) before the implementation
starts.

### Issues

- **One issue per roadmap item**, in its milestone, with the item's text as
  the description and a shorter title. For example, "Leitner scheduling" in
  M4, with the box intervals in the description.
- **The items already done** (the first two in M0) become closed issues that
  link to their PRs ([lafeingcrokodil/flashcards#32](https://github.com/lafeingcrokodil/flashcards/pull/32)
  and [lafeingcrokodil/flashcards#35](https://github.com/lafeingcrokodil/flashcards/pull/35)),
  so that M0's progress counts them. "Before M0" (documenting the workflow)
  isn't carried over, since it's in `AGENTS.md`'s history.
- **The three M0 items from the discussion on #38** (`govulncheck`, the
  monthly version check, and the check for the "Before merging" checklist)
  become issues in M0, after Dependabot.
- **"Later ideas"** (a staging environment, an installable PWA) become
  issues without a milestone, with an `idea` label.
- **"Open questions"** become issues with a `question` label, in the
  milestone where they need an answer: the custom domain in M2, and where
  packs and passages come from in M3.
- **Not assigned to anyone:** they're planned work, which either of us can
  pick up, unlike `follow-up` issues, which are assigned to you.

🟡 **Order within a milestone:** milestones don't keep their issues in a
chosen order. I'll create the issues in roadmap order, so that sorting a
milestone by "Oldest" shows them in that order. When the order changes, it's
a comment on the issue rather than an edit.

- Rejected for now: a Project board, which can order items, but is one more
  thing to keep up to date.
- Rejected: a tracking issue per milestone, with the items as sub-issues.
  Sub-issues keep their order, and I could create everything myself. But a
  milestone is GitHub's own way to group issues and PRs and show their
  progress, and PRs can be in milestones too, while they can't be
  sub-issues.

### Labels

- `follow-up` (from #38), `idea` and `question`.
- 🔴 I don't know whether creating an issue with a label that doesn't exist
  yet creates the label. I'll try it with the first issue. If it doesn't, I'll
  stop and ask you to create the three labels too (Issues → Labels → New
  label).

### `docs/architecture.md`

The introduction and the "Architecture" section of `docs/roadmap.md`, moved
unchanged. `docs/roadmap.md` is deleted rather than kept as a pointer to the
milestones, since a link in the README and `AGENTS.md` does that job.

The README gets a link to the milestones and to `docs/architecture.md`.

### `AGENTS.md`

- "Keep `docs/roadmap.md` up to date" becomes: planned work is tracked as
  issues in milestones. A PR that finishes an issue says
  `Closes lafeingcrokodil/flashcards#N` in its description, so that merging
  it closes the issue. New planned work gets a new issue in the right
  milestone.
- The "Follow-up issues" section (from #38) stops distinguishing between the
  roadmap and issues: anything left to do is an issue, and `follow-up` marks
  the ones that are your tasks after a PR.
- The plan steps say that a plan names the issue it's for.

🟡 **Departure from diary,** which keeps a `docs/roadmap.md`. A file there
works too, but it splits tasks between the file and issues, which is the
problem this change is for.

### Order of the work

1. You create the eight milestones (after this plan is approved).
2. I create the issues, close the done ones, and check the milestones.
3. The PR's commits move the architecture to `docs/architecture.md`, delete
   `docs/roadmap.md`, and update `AGENTS.md` and the README.

Open PRs that refer to the roadmap ([lafeingcrokodil/flashcards#37](https://github.com/lafeingcrokodil/flashcards/pull/37)
ticks its item; [lafeingcrokodil/flashcards#38](https://github.com/lafeingcrokodil/flashcards/pull/38)
mentions it in `AGENTS.md`) are adjusted to say `Closes ...` instead, in
whichever order they're merged.

## Risks

- 🔴 **The issues don't match the roadmap** (an item missing, in the wrong
  milestone, or with its details lost): no check catches it, and the issues
  aren't part of the PR's diff. The PR description lists every issue next to
  the roadmap item it came from, so they can be compared, and the deleted
  `docs/roadmap.md` stays in the diff for reference.
- **The issues exist but the PR isn't merged** (or the other way round):
  until both are done, the roadmap and the issues would disagree. The issues
  can be closed as "not planned" to undo it, and the PR can be reverted.
- **Planned work loses its history of review:** changes to issues and
  milestones aren't reviewed in PRs, as edits to `docs/roadmap.md` were.
  Accepted: the issues' own history shows who changed what.

## Close review

- 🔴 **The created issues and milestones**, against the deleted roadmap,
  using the list in the PR description.
- **`AGENTS.md`**: the rules for when to open, close and label issues, since
  later sessions follow them and no check does.
