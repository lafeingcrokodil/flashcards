// What the version checks share: comparing versions, picking the newest
// release that's old enough, and reading and fetching versions.
const fs = require('fs')

const minAgeDays = 14

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

// isBehind says whether current is older than newest, the result of
// newestEligible.
function isBehind(current, newest) {
  return Boolean(newest) && compareVersions(current, newest.version) < 0
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

// day formats a date as YYYY-MM-DD.
function day(date) {
  return date.toISOString().slice(0, 10)
}

module.exports = {
  minAgeDays,
  parseVersion,
  compareVersions,
  newestEligible,
  isBehind,
  readVersion,
  fetchText,
  day,
}
