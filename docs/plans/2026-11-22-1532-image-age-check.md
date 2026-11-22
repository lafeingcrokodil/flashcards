# A CI check that the base image digests are at least 14 days old

For lafeingcrokodil/flashcards#110.

Markers: 🟡 a judgement call where you might prefer the alternative; 🔴
risky or uncertain.

## Goal

A CI job, `image-age`, that fails while a pull request adds a base image
digest to the `Dockerfile` that was published less than 14 days ago, and says
when it will be old enough. A Dependabot PR such as
lafeingcrokodil/flashcards#108 then stays red until it can be merged, instead
of each digest being dated by hand.

Left for later: re-running the job automatically once a digest is old
enough. Until then, it's re-run by hand (or by Dependabot rebasing the PR);
with monthly Dependabot updates, that's rarely more than one PR at a time.

## Design

### What's checked: the digests that the PR adds

The job compares the `FROM` lines of the PR's `Dockerfile` with those of the
base branch's, and checks only the digests that are new. A digest that's
already on `main` was checked when it was added (or accepted as an
exception), so it isn't checked again. That means:

- A PR that doesn't change a `FROM` line passes without any network access.
- `main` doesn't turn red when a digest that's on it can't be dated any more,
  for example because the registry has dropped its history.
- The job runs on pull requests only. On pushes to `main`, it's skipped.

🟡 **On every PR, not only Dependabot's:** a base image bumped by hand gets
the same check, and other PRs pass it straight away.

- Rejected: only Dependabot's PRs (`github.actor == 'dependabot[bot]'`). It
  saves nothing, since the check is free for PRs that don't change a `FROM`
  line, and a bump by hand would be unchecked.

### 🟡 Which date: the `linux/amd64` image's

Each `FROM` line pins an image index (a list of images, one per platform),
and the build only uses its `linux/amd64` image. The check dates that image,
not the index: an official image's index is pushed again from time to time
without changing its `linux/amd64` image (for example
`golang:1.27.4-trixie`'s, on 2026-11-15), and that doesn't bring in anything
new that the build runs.

- Rejected: dating the index. A re-pushed index would make an unchanged
  image look new and hold up the PR for nothing.

### Where the dates come from

| Registry | `linux/amd64` image from | Date from |
| --- | --- | --- |
| `gcr.io` (distroless) | the index's manifest, fetched from the registry | the registry's tag list (`/v2/<image>/tags/list`), whose `timeUploadedMs` it records for every digest it holds, old ones included |
| Docker Hub official images (`node`, `golang`) | [docker-library/repo-info](https://github.com/docker-library/repo-info)'s `repos/<image>/remote/<tag>.md`, which lists the index's digest and each platform's | the commit date of the first version of that file that lists the `linux/amd64` image |

For distroless, the tag list gives `25ac8f0…` (on `main`) and its
`linux/amd64` image `336e53e…`, both uploaded on 2026-10-31, matching what
lafeingcrokodil/flashcards#108's comment found by hand.

For Docker Hub, repo-info rather than Hub's tags API
(`hub.docker.com/v2/repositories/library/<image>/tags/<tag>`), because the
API only knows a tag's current digest. A Dependabot PR that waits 14 days
will often find the tag rebuilt by then, so its digest is no longer current,
and the API can't date it. repo-info's history keeps every digest a tag has
had, even after the tag is no longer maintained (`node:26.13.0-trixie-slim`'s
file is gone from its `master`, but not from its history).

repo-info's bot commits a few minutes after each push: for
`golang:1.27.4-trixie`, Hub says its `linux/amd64` image was pushed at
18:30 UTC on 2026-11-10, and repo-info first lists it at 18:47. So the date
is slightly late, which only makes the check a little stricter.

🟡 **How the history is read:** a shallow, blob-less clone of repo-info
covering the last 60 days (`git clone --shallow-since=… --filter=blob:none
--no-checkout`, about a second), then `git log` and `git show` on the one
file. If the image is already listed in the oldest version of the file in
the clone, it's at least 60 days old, which is old enough.

- Rejected: GitHub's API for the file's commits. It would work in CI, but I
  can't try it from this environment, whose GitHub API access is limited to
  this repository, whereas the clone works here.
- Rejected: Docker Hub's registry, whose image configs have a `created`
  date. It's set by whoever builds the image, rather than recorded by the
  registry or a third party, and pulling manifests anonymously is rate
  limited (it already is from this environment).

### 🟡 When a digest can't be dated: fail

If the digest isn't in the registry's or repo-info's records (for example an
image from another registry, a repo-info file that doesn't exist, or a
format change), the job fails and says so. An exception (below) is the way
past it.

- Rejected: a warning. It's easy to miss on a green PR, and then the check
  would go quiet exactly when it's needed.

A `FROM` line that the check can't parse (such as one without a digest, or
with `--platform`) fails it too, so that a new form of `FROM` line can't slip
past.

### 🟡 Exceptions: a comment in the `Dockerfile` naming the digest

A digest that's too young, or can't be dated, is accepted if the line
directly above its `FROM` line is a comment like:

```dockerfile
# Image age exception for sha256:f18c3d73…: 12 days old, and there's no older digest for Go 1.27.4.
```

with the full digest, followed by the reason. The job then passes with a
warning that names the exception. Naming the digest means that when the
image is bumped again, the exception no longer applies; the job then fails
until the stale comment is removed, so that it doesn't mislead readers.

The `golang` digest that lafeingcrokodil/flashcards#104 made an exception
for is already on `main`, so it doesn't need a comment.

- Rejected: a PR label. It's quicker to add, but it isn't kept anywhere you'd
  see it once the PR is merged, and it would apply to every digest in the
  PR rather than one.
- Rejected: a warning for a stale exception, rather than failing. It's one
  more line to remove by hand in a Dependabot PR, but exceptions should be
  rare.

### The job and the script

- **`image-age` in `ci.yml`:** checks out the PR, fetches the base commit
  (`github.event.pull_request.base.sha`) to read its `Dockerfile`, and runs
  `node .github/scripts/check-image-ages.js`. It uses the runner's Node (only
  `fetch`, `fs` and `child_process`), so there's no `setup-node` or
  dependencies. It has the workflow's read-only permissions, and needs no
  token, since the registries, the tag list and repo-info are public.
- **The messages:** an `::error` annotation on the `FROM` line, such as
  "`golang:1.27.4-trixie@sha256:…` (`linux/amd64` image `sha256:2cd7e1e…`)
  was published on 2026-11-10, so it's 12 days old. It's 14 days old on
  2026-11-24: re-run this job then." For an image that can't be dated, it
  says why and how to add an exception.
- 🟡 **Committed tests,** `.github/scripts/check-image-ages.test.js`, with
  Node's built-in test runner (`node --test`), run by the same job before
  the check. They cover the parsing, the comparison with the base, the
  exceptions and the age logic, with fake registry and repo-info data, so
  they need no network. This departs from `check-versions.js`, which has no
  committed tests, because there's more logic here, and a mistake would let
  a young digest through without anyone noticing.
- **The README's "Automated checks" table** mentions the new check in
  `ci.yml`'s row, and AGENTS.md's "Dependencies" section says that CI checks
  the base images' age, and how to make an exception.

## Risks

- **A young digest passes** (for example because the wrong image's date is
  used, or a `FROM` line is skipped): the tests, which include a young digest,
  a re-pushed index with an old `linux/amd64` image, and unparseable `FROM`
  lines. Before marking the PR as ready, I'll also run the script locally
  against lafeingcrokodil/flashcards#108's `Dockerfile`, which should fail
  until 2026-11-26.
- **A source changes its format or disappears** (the tag list, the manifest,
  repo-info's files): the digest can't be dated, so the job fails rather
  than passing.
- 🔴 **repo-info's or the registry's dates are wrong.** The check trusts
  them, and can't verify them. They're recorded by the registry and by
  docker-library's bot, not by whoever builds the image, and I've only
  compared them with Hub's dates for the three images above.
- **The job fails while a digest is old enough** (for example a network
  error): it says why, and can be re-run.
- **A PR goes green and nobody notices it's mergeable:** the job is only
  re-run by hand. Low harm: it only delays an update.

## Close review

- **`.github/workflows/ci.yml`'s `image-age` job:** a CI change. It's
  read-only and runs no code from the PR other than the script, but a
  mistake in when it runs (such as skipping it on PRs) wouldn't be caught.
- **The exception rule** in `check-image-ages.js`: it decides when a young
  digest is let through.

The rest is covered by the tests, or fails visibly.
