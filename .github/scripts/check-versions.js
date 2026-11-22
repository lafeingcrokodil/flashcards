// Checks whether Go, golangci-lint, Node and pnpm are behind the newest
// release that's at least minAgeDays old, and reports each as a GitHub issue
// with report-issue.js: opened while it's out of date, closed once it isn't.
//
// The versions in use are read from the repository's files. The releases and
// their dates come from the Go module proxy for Go and golangci-lint, from
// nodejs.org for Node, and from the npm registry for pnpm. It's meant to be
// run with actions/github-script, from the repository's root.
const fs = require('fs')
const report = require('./report-issue.js')

const minAgeDays = 14
const proxy = 'https://proxy.golang.org'
const goModule = 'golang.org/toolchain'
const lintModule = 'github.com/golangci/golangci-lint/v2'
const nodeIndex = 'https://nodejs.org/dist/index.json'
const pnpmPackage = 'https://registry.npmjs.org/pnpm'

// parseVersion returns a version's numeric parts ("1.27.4" or "v2.15.0" gives
// [1, 27, 4] or [2, 15, 0]), or null for anything else, including
// pre-releases such as "1.28rc1", "v2.16.0-rc.1" or "v27.0.0-alpha.0.1.1".
function parseVersion(version) {
  const match = /^v?(\d+)\.(\d+)(?:\.(\d+))?$/.exec(version)
  return match ? match.slice(1).map((part) => Number(part ?? 0)) : null
}

// compareVersions returns a negative number if a is older than b, a positive
// one if it's newer, and 0 if they're the same.
function compareVersions(a, b) {
  const [pa, pb] = [parseVersion(a), parseVersion(b)]
  for (let i = 0; i < 3; i++) {
    if (pa[i] !== pb[i]) return pa[i] - pb[i]
  }
  return 0
}

// newestEligible returns the newest release that's at least minAgeDays old,
// as {version, date}, or null if there's none. releases are stable versions,
// and dateOf returns a release's date (a Date), fetching it if needed, so
// that only the newest few releases' dates are fetched.
async function newestEligible(releases, dateOf, now) {
  const newestFirst = [...releases].sort((a, b) => compareVersions(b, a))
  for (const version of newestFirst) {
    const date = await dateOf(version)
    if (now - date >= minAgeDays * 24 * 60 * 60 * 1000) return { version, date }
  }
  return null
}

// readVersion returns the first capture of pattern in the file at path, and
// throws if there's none, so that a change to the file's format fails the
// check rather than making it report that everything is up to date.
function readVersion(path, pattern) {
  const match = pattern.exec(fs.readFileSync(path, 'utf8'))
  if (!match) throw new Error(`No version matching ${pattern} in ${path}`)
  return match[1]
}

async function fetchText(url) {
  const response = await fetch(url)
  if (!response.ok) throw new Error(`${url}: ${response.status}`)
  return response.text()
}

async function proxyDate(module, version) {
  const info = JSON.parse(await fetchText(`${proxy}/${module}/@v/${version}.info`))
  return new Date(info.Time)
}

// goReleases lists Go's stable releases (e.g. "1.27.4"). The toolchain
// module has a version per release and platform, so one platform's are used.
async function goReleases() {
  const list = await fetchText(`${proxy}/${goModule}/@v/list`)
  return list
    .split('\n')
    .map((line) => /^v0\.0\.1-go(.+)\.linux-amd64$/.exec(line.trim()))
    .filter((match) => match && parseVersion(match[1]))
    .map((match) => match[1])
}

async function lintReleases() {
  const list = await fetchText(`${proxy}/${lintModule}/@v/list`)
  return list
    .split('\n')
    .map((line) => line.trim())
    .filter((version) => parseVersion(version))
}

// nodeReleases lists Node's stable releases as {version, date, lts}, where
// version is like "26.13.1" and lts says whether it's an LTS release.
async function nodeReleases() {
  const index = JSON.parse(await fetchText(nodeIndex))
  return index
    .map((release) => ({
      version: release.version.replace(/^v/, ''),
      date: new Date(release.date),
      lts: release.lts !== false,
    }))
    .filter((release) => parseVersion(release.version))
}

// nodeCandidates returns the Node releases that may be updated to from
// current: any release in its major version, and LTS releases in newer
// ones, so that a new major only counts once its line has become LTS.
function nodeCandidates(releases, current) {
  const major = parseVersion(current)[0]
  return releases.filter((release) => {
    const releaseMajor = parseVersion(release.version)[0]
    return releaseMajor === major || (releaseMajor > major && release.lts)
  })
}

// pnpmReleases returns pnpm's stable versions that are no newer than its
// "latest" dist-tag, since newer ones are only tagged "next", and the time
// field of its registry document, which maps each version to its publish
// date.
async function pnpmReleases() {
  const doc = JSON.parse(await fetchText(pnpmPackage))
  const latest = doc['dist-tags'].latest
  const versions = Object.keys(doc.time).filter(
    (version) => parseVersion(version) && compareVersions(version, latest) <= 0,
  )
  return { versions, time: doc.time }
}

function day(date) {
  return date.toISOString().slice(0, 10)
}

// goBody describes how Go is out of date, or returns '' if it isn't.
function goBody(current, newest) {
  if (!newest || compareVersions(current, newest.version) >= 0) return ''
  return [
    `\`go.mod\` requires Go ${current}, but Go ${newest.version} (released on ${day(newest.date)}) is the newest release that's at least ${minAgeDays} days old.`,
    '',
    `To update, run \`go mod edit -go=${newest.version}\`, and check the [release notes](https://go.dev/doc/devel/release) for anything that affects the code.`,
  ].join('\n')
}

// lintBody describes how golangci-lint is out of date, or how the README
// disagrees with CI's version (ci, which current only differs from when it's
// overridden), or returns '' if neither.
function lintBody(current, ci, readme, newest) {
  const lines = []
  if (newest && compareVersions(current, newest.version) < 0) {
    lines.push(
      `CI uses golangci-lint ${current}, but ${newest.version} (released on ${day(newest.date)}) is the newest release that's at least ${minAgeDays} days old.`,
      '',
      `To update, change the \`version\` of the \`lint\` job in \`.github/workflows/ci.yml\` and the version in \`README.md\`, and check the [changelog](https://golangci-lint.run/docs/product/changelog/) for new linters to configure.`,
    )
  }
  if (readme !== ci) {
    if (lines.length) lines.push('')
    lines.push(
      `\`README.md\` says golangci-lint ${readme}, but CI uses ${ci}, so the README should say ${ci} too.`,
    )
  }
  return lines.join('\n')
}

// nodeBody describes how Node is out of date, or returns '' if it isn't.
function nodeBody(current, newest) {
  if (!newest || compareVersions(current, newest.version) >= 0) return ''
  const lines = [
    `\`web/.nvmrc\` pins Node ${current}, but Node ${newest.version} (released on ${day(newest.date)}) is the newest release that's at least ${minAgeDays} days old, among Node ${parseVersion(current)[0]}'s releases and newer major versions' LTS releases.`,
    '',
    `To update, change \`web/.nvmrc\`, and check the [release notes](https://nodejs.org/en/blog/release/v${newest.version}) for anything that affects the front end.`,
  ]
  const [major, newMajor] = [parseVersion(current)[0], parseVersion(newest.version)[0]]
  if (newMajor !== major) {
    lines.push(
      '',
      `It's a new major version, so change \`engines.node\` in \`web/package.json\` to \`>=${newMajor}\` too.`,
    )
  }
  return lines.join('\n')
}

// pnpmBody describes how pnpm is out of date, or returns '' if it isn't.
function pnpmBody(current, newest) {
  if (!newest || compareVersions(current, newest.version) >= 0) return ''
  return [
    `\`packageManager\` in \`web/package.json\` says pnpm ${current}, but pnpm ${newest.version} (released on ${day(newest.date)}) is the newest release that's at least ${minAgeDays} days old and no newer than the \`latest\` tag.`,
    '',
    `To update, change \`packageManager\` in \`web/package.json\`, and check the [release notes](https://github.com/pnpm/pnpm/releases/tag/v${newest.version}) for anything that affects the front end.`,
  ].join('\n')
}

// main runs the check. Its overrides replace the versions read from the
// files, so that a manual run can test what happens when one is out of date.
async function main({ github, context, overrides = {}, now = new Date() }) {
  const currentGo = overrides.go || readVersion('go.mod', /^go (\S+)$/m)
  const ciLint = readVersion(
    '.github/workflows/ci.yml',
    /golangci-lint-action@.*\n(?:.*\n)*?\s*version: (\S+)/,
  )
  const currentLint = overrides.lint || ciLint
  const readmeLint = readVersion('README.md', /golangci-lint\]\(.*\) (v\S+)/)

  const currentNode = overrides.node || readVersion('web/.nvmrc', /^v?(\d+\.\d+\.\d+)\s*$/)
  const currentPnpm =
    overrides.pnpm ||
    readVersion('web/package.json', /"packageManager": "pnpm@(\d+\.\d+\.\d+)(?:\+[^"]*)?"/)

  // Each tool is reported before the next one's releases are fetched, so
  // that one source being unreachable doesn't stop the reports before it.
  const newestGo = await newestEligible(
    await goReleases(),
    (version) => proxyDate(goModule, `v0.0.1-go${version}.linux-amd64`),
    now,
  )
  console.log(`Go: ${currentGo} in use, newest eligible ${newestGo?.version}`)
  await report({ github, context, title: 'Go is out of date', body: goBody(currentGo, newestGo) })

  const newestLint = await newestEligible(
    await lintReleases(),
    (version) => proxyDate(lintModule, version),
    now,
  )
  console.log(`golangci-lint: ${currentLint} in use (README: ${readmeLint}), newest eligible ${newestLint?.version}`)
  await report({
    github,
    context,
    title: 'golangci-lint is out of date',
    body: lintBody(currentLint, ciLint, readmeLint, newestLint),
  })

  const node = nodeCandidates(await nodeReleases(), currentNode)
  const nodeDates = new Map(node.map((release) => [release.version, release.date]))
  const newestNode = await newestEligible(
    node.map((release) => release.version),
    async (version) => nodeDates.get(version),
    now,
  )
  console.log(`Node: ${currentNode} in use, newest eligible ${newestNode?.version}`)
  await report({ github, context, title: 'Node is out of date', body: nodeBody(currentNode, newestNode) })

  const pnpm = await pnpmReleases()
  const newestPnpm = await newestEligible(pnpm.versions, async (version) => new Date(pnpm.time[version]), now)
  console.log(`pnpm: ${currentPnpm} in use, newest eligible ${newestPnpm?.version}`)
  await report({ github, context, title: 'pnpm is out of date', body: pnpmBody(currentPnpm, newestPnpm) })
}

module.exports = main
// For checking the logic without the network or GitHub.
module.exports.internals = {
  parseVersion,
  compareVersions,
  newestEligible,
  nodeCandidates,
  goBody,
  lintBody,
  nodeBody,
  pnpmBody,
  readVersion,
}
