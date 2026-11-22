# A CI check that the distroless digest is at least 14 days old

For lafeingcrokodil/flashcards#110.

Markers: 🟡 a judgement call where you might prefer the alternative; 🔴
risky or uncertain.

## Goal

A CI job, `image-age`, that fails while the `Dockerfile` pins a `gcr.io`
base image digest that was uploaded less than 14 days ago, and says when it
will be old enough. A Dependabot PR such as lafeingcrokodil/flashcards#108
then stays red until it can be merged, instead of the digest being dated by
hand.

Left for later:

- Docker Hub images (`node`, `golang`). Dependabot's cooldown already holds
  their updates back (see below), so the check leaves them to it for now.
- Re-running the job automatically once a digest is old enough. Until then,
  it's re-run by hand (or by Dependabot rebasing the PR).

## Design

### 🟡 Only `gcr.io`: Docker Hub is left to Dependabot

Since dependabot/dependabot-core#16143 (September 2026), Dependabot's Docker
cooldown dates updates with Docker Hub's `tag_last_pushed`, so a `node` or
`golang` update waits 14 days. Other registries have no date, so Dependabot
opens the PR straight away with a warning, as for #108's distroless digest.
So the gap is `gcr.io`, and that's what the check covers.

- 🔴 **Not verified:** whether Dependabot's cooldown also holds back
  *digest-only* updates of Docker Hub tags. The PR's original description
  said so, but the final docs don't. The next such PR will show it: if it's
  opened with a digest pushed less than 14 days before, I'll open an issue to
  extend the check to Docker Hub.
- A Docker Hub image bumped by hand isn't checked either way.
- Rejected: covering Docker Hub too, with
  [docker-library/repo-info](https://github.com/docker-library/repo-info)'s
  history, as the first version of this plan did. It works, but it's most of
  the plan's complexity, for a case that Dependabot covers.
- Rejected: [digestabot](https://github.com/chainguard-dev/digestabot) or
  [dockerfile-pin](https://github.com/azu/dockerfile-pin), which both have a
  minimum age. They date an image by its config's `created` field, which is
  always 1970 for distroless (it's built reproducibly), and which whoever
  builds the image can set to anything.

### What's checked

Every `FROM` line whose image is on `gcr.io`, on every PR and push to `main`,
like the `image` job's version check. The registry's tag list
(`https://gcr.io/v2/<image>/tags/list`) has a `timeUploadedMs` for every
digest it holds, old ones included, so a digest that's on `main` can still be
dated, and there's no need to compare with the base branch. For today's
`25ac8f0…`, it says 2026-10-31, as #108's comment found by hand.

The date is the pinned digest's (the image index's), not its `linux/amd64`
image's. distroless pushes them together (`25ac8f0…` and its `linux/amd64`
image `336e53e…` were both uploaded on 2026-10-31), so dating the index is
simpler and no less strict.

`FROM` lines for other registries are skipped, except that the check fails on
a `FROM` line it can't parse (such as one without a digest, or with
`--platform`), so that a new form of `FROM` line can't slip past.

### 🟡 When a digest can't be dated: fail

If the digest isn't in the tag list (or the list can't be fetched, or its
format has changed), the job fails and says so. An exception (below) is the
way past it.

- Rejected: a warning. It's easy to miss on a green PR.

### 🟡 Exceptions: a comment in the `Dockerfile` naming the digest

A digest that's too young, or can't be dated, is accepted if the line
directly above its `FROM` line is a comment like:

```dockerfile
# Image age exception for sha256:aafe9936…: <the reason>
```

with the full digest. The job then passes with a warning. Naming the digest
means that when the image is bumped again, the exception no longer applies;
the job then fails until the stale comment is removed, so that it doesn't
mislead readers.

- Rejected: a PR label. It isn't kept anywhere you'd see it once the PR is
  merged, and would apply to every digest in the PR.
- Rejected: only a warning for a stale exception. It's one more line to
  remove by hand in a Dependabot PR, but exceptions should be rare.

### The job and the script

- **`image-age` in `ci.yml`:** checks out the repository and runs
  `node .github/scripts/check-image-ages.js`. It uses the runner's Node (only
  `fetch` and `fs`), so there's no `setup-node` or dependencies, and the
  workflow's read-only permissions. No token: the tag list is public.
- **The messages:** an `::error` annotation on the `FROM` line, such as
  "`gcr.io/distroless/static-debian13:nonroot@sha256:aafe993…` was uploaded
  on 2026-11-12, so it's 10 days old. It's 14 days old on 2026-11-26: re-run
  this job then." For a digest that can't be dated, it says why and how to
  add an exception.
- **lafeingcrokodil/flashcards#105:** its `.github/scripts/versions/common.js`
  has `minAgeDays`, `fetchText` and `day`. If #105 is merged first, the new
  script reuses them; otherwise it has its own, and #105 can switch to them.
  `check-image-ages.js` stays outside `versions/`, since it's a check that
  gates PRs rather than a monthly report.
- 🟡 **Committed tests,** `.github/scripts/check-image-ages.test.js`, with
  Node's built-in test runner (`node --test`), run by the same job before
  the check, with a fake tag list, so they need no network. They cover the
  parsing, the exceptions and the age. This departs from the version checks,
  which have no committed tests, because this one gates merging: a mistake
  would let a young digest through without anyone noticing.
- **The README's "Automated checks" table** mentions the check in `ci.yml`'s
  row, and AGENTS.md's "Dependencies" section says what CI checks and how to
  make an exception.

## Risks

- **A young digest passes** (for example a `FROM` line is skipped, or the
  wrong digest's date is used): the tests, which include a young digest,
  another image's digest in the tag list, and unparseable `FROM` lines.
  Before marking the PR as ready, I'll also run the script against #108's
  `Dockerfile`, which should fail until 2026-11-26.
- **The tag list changes its format, or drops old digests:** the digest
  can't be dated, so the job fails rather than passing. If it drops old
  digests, `main` turns red too, which is visible.
- **A Docker Hub digest under 14 days old gets in:** see the 🔴 point above.
- **The job fails while the digest is old enough** (for example a network
  error): it says why, and can be re-run.

## Close review

- **`.github/workflows/ci.yml`'s `image-age` job:** a CI change. It's
  read-only, but a mistake in when it runs wouldn't be caught.
- **The exception rule** in `check-image-ages.js`: it decides when a young
  digest is let through.

The rest is covered by the tests, or fails visibly.
