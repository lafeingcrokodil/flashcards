// Checks whether Go and golangci-lint are behind the newest release that's at
// least minAgeDays old, and reports each as a GitHub issue with
// report-issue.js: opened while it's out of date, closed once it isn't.
//
// The versions in use are read from the repository's files, and the releases
// and their dates from the Go module proxy, which lists both tools. It's meant
// to be run with actions/github-script, from the repository's root.
const fs = require('fs')
const report = require('./report-issue.js')

const minAgeDays = 14
const proxy = 'https://proxy.golang.org'
const goModule = 'golang.org/toolchain'
const lintModule = 'github.com/golangci/golangci-lint/v2'

// parseVersion returns a version's numeric parts ("1.27.4" or "v2.15.0" gives
// [1, 27, 4] or [2, 15, 0]), or null for anything else, including
// pre-releases such as "1.28rc1" or "v2.16.0-rc.1".
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

  const newestGo = await newestEligible(
    await goReleases(),
    (version) => proxyDate(goModule, `v0.0.1-go${version}.linux-amd64`),
    now,
  )
  const newestLint = await newestEligible(
    await lintReleases(),
    (version) => proxyDate(lintModule, version),
    now,
  )
  console.log(`Go: ${currentGo} in use, newest eligible ${newestGo?.version}`)
  console.log(`golangci-lint: ${currentLint} in use (README: ${readmeLint}), newest eligible ${newestLint?.version}`)

  await report({ github, context, title: 'Go is out of date', body: goBody(currentGo, newestGo) })
  await report({
    github,
    context,
    title: 'golangci-lint is out of date',
    body: lintBody(currentLint, ciLint, readmeLint, newestLint),
  })
}

module.exports = main
// For checking the logic without the network or GitHub.
module.exports.internals = { parseVersion, compareVersions, newestEligible, goBody, lintBody, readVersion }
