# Check Node and pnpm versions monthly

For lafeingcrokodil/flashcards#99.

Markers: 🟡 a judgement call where you might prefer the alternative; 🔴
risky or uncertain.

## Goal

Dependabot doesn't update the front end's Node version (`web/.nvmrc`), and
it isn't confirmed whether it updates pnpm's (`packageManager` in
`web/package.json`). The monthly version check from
lafeingcrokodil/flashcards#44 (`.github/workflows/versions.yml` and
`.github/scripts/check-versions.js`), which does this for Go and
golangci-lint, also reports Node and pnpm when they're behind the newest
release that's at least 14 days old: one issue each, "Node is out of date"
and "pnpm is out of date", opened and closed with `report-issue.js` like
the existing two.

Left for later: updating the versions automatically, as for Go and
golangci-lint. That's the idea in lafeingcrokodil/flashcards#93, whose scope
now includes Node and pnpm too.

## Design

### What's compared

| Tool | Current version from | Newest release from |
| --- | --- | --- |
| Node | `web/.nvmrc` (`26.13.0`) | `https://nodejs.org/dist/index.json`, which lists every release with its date and LTS name |
| pnpm | `packageManager` in `web/package.json` (`pnpm@12.18.0`, with or without a `+sha…` hash suffix) | the npm registry's document for `pnpm` (`https://registry.npmjs.org/pnpm`), whose `time` field has each version's publish date |

As for Go and golangci-lint, pre-releases are excluded (Node's
`v27.0.0-alpha…`, pnpm's `-rc` and the like), and the rest of the
comparison (`parseVersion`, `compareVersions`, `newestEligible`) is
reused. Both sources were reachable from this environment, and are
public, so the workflow needs no new permissions or secrets.

Today both are up to date: Node 26.13.1 (2026-11-10) and pnpm 12.19.0
(2026-11-10) and later are less than 14 days old.

### 🟡 Node: the same major, or a newer LTS major

A Node release counts if it's in the major version in use, or in a newer
major that has had an LTS release. So a new major's Current releases
(before it becomes LTS, about six months after its first release) don't
count, but once the line is LTS, its LTS releases that are at least 14
days old do, and the issue then suggests moving to it. For a newer major,
only LTS releases count (`lts` isn't `false` in `index.json`); in the major
in use, every release counts, so that a project on a Current line still
gets its patches.

- Rejected: only the major in use, as the issue suggests. Simpler, but it
  would never say when a new LTS line is ready, and the one in use would
  quietly reach its end of life.
- Rejected: any release, which would nag about each new major from its
  first release, before it's LTS.

When the major changes, `engines.node` (`>=26`) in `web/package.json`
should change with it, so the issue says so.

### 🟡 pnpm: no newer than the `latest` tag

A pnpm version counts if it's at least 14 days old and no newer than the
registry's `latest` dist-tag. pnpm publishes new versions under the `next`
tag first and moves `latest` to them later (today `latest` is 12.21.0, and
12.22.0 and 12.23.0 are only `next`), so this keeps the check to versions
that pnpm itself calls stable. Older majors' maintenance releases (11.x)
never count, since they're older than the version in use.

- Rejected: any stable-looking version, which would count `next` versions
  that pnpm hasn't promoted yet, once they're 14 days old.

The full registry document is about 6 MB (the abbreviated one, which npm
clients use, has no dates). That's fine for a monthly run.

If Dependabot does update `packageManager` (which
lafeingcrokodil/flashcards#98 left unconfirmed), it and this check overlap:
the issue is closed by the next run after Dependabot's PR is merged.
Harmless, so no need to find out first.

### Workflow

`versions.yml` gets two more optional inputs for manual runs, `node` and
`pnpm`, which replace the versions read from the files, as the existing
`go` and `golangci-lint` inputs do. Its comment about what Dependabot
doesn't update mentions Node and pnpm too. Nothing else changes: same
schedule, same permissions.

### One module per tool

With four tools, `check-versions.js` would read four versions, fetch four
sources and make four reports in one `main`, and each new tool (Terraform,
say) would make that worse. So the script is split by tool (decided in
review):

- `.github/scripts/versions/common.js`: what all the tools share
  (`minAgeDays`, `parseVersion`, `compareVersions`, `newestEligible`,
  `isBehind`, `readVersion`, `fetchText`, `day`).
- `versions/go.js`, `golangci-lint.js`, `node.js` and `pnpm.js`: each
  exports `{ key, title, check({ override, now }) }`, where `check` reads
  the version in use, fetches the releases, logs, and returns the issue's
  body (`''` if up to date), and `key` names its entry in the workflow's
  overrides.
- `check-versions.js`: only `main`, a loop over the tools that calls
  `check`, then `report`, so that each tool is reported before the next
  one's releases are fetched. The workflow's call doesn't change.

Go and golangci-lint behave as before.

- Rejected: one file with a table of tools. That makes `main` a loop too,
  but the file would still grow with every tool.

### Testing

As for #44, there are no committed tests for these scripts. Instead:

- Before marking the PR as ready, I'll run the new logic locally against
  fake inputs (an older Node 26 patch, a newer major without LTS releases,
  a newer major with an old enough LTS release, a pnpm version above
  `latest`, a `packageManager` with a hash suffix), and the whole check
  against the real sources with a stubbed `report`.
- After merging, a manual run with `node` set to `26.11.0` and `pnpm` to
  `12.10.0` opens both issues, and a run without inputs closes them.

## Risks

- **A version is compared or filtered wrongly** (e.g. a Current release of a
  new major counted, or a `next` pnpm version): covered by the local run with
  fake inputs before merging, not by a committed test.
- **The check reads the wrong version from a file** (e.g. `.nvmrc` with a `v`
  prefix, or `packageManager` naming another tool): if the format doesn't
  match, the run fails rather than reporting "up to date", as for the
  existing tools. A leading `v` in `.nvmrc` is accepted.
- **nodejs.org or the npm registry is unreachable or changes its format:**
  the run fails, which sends a notification. The Go and golangci-lint checks
  run first, so their issues are still updated.
- **Node's `lts` field changes meaning** (e.g. a line's releases from before
  it became LTS are later marked as LTS): only affects when the issue first
  suggests a new major, which does little harm.

## Close review

None. The workflow's permissions and triggers are unchanged (only two inputs
are added), and a wrong comparison only means a wrong or missing reminder,
which is easy to fix and does little harm. The closest is the Node rule
above, which is a judgement call rather than a risk.
