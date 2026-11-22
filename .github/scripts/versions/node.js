// Checks Node: the version in web/.nvmrc, against the releases listed by
// nodejs.org.
const { minAgeDays, parseVersion, newestEligible, isBehind, readVersion, fetchText, day, dockerfileLine } = require('./common.js')

const index = 'https://nodejs.org/dist/index.json'

// releases lists Node's stable releases as {version, date, lts}, where
// version is like "26.13.1" and lts says whether it's an LTS release.
async function releases() {
  return JSON.parse(await fetchText(index))
    .map((release) => ({
      version: release.version.replace(/^v/, ''),
      date: new Date(release.date),
      lts: release.lts !== false,
    }))
    .filter((release) => parseVersion(release.version))
}

// candidates returns the releases that may be updated to from current: any
// release in its major version, and LTS releases in newer ones, so that a
// new major only counts once its line has become LTS.
function candidates(releases, current) {
  const major = parseVersion(current)[0]
  return releases.filter((release) => {
    const releaseMajor = parseVersion(release.version)[0]
    return releaseMajor === major || (releaseMajor > major && release.lts)
  })
}

// newest returns the newest candidate that's old enough, as {version, date},
// or null if there's none.
function newest(releases, current, now) {
  const eligible = candidates(releases, current)
  const dates = new Map(eligible.map((release) => [release.version, release.date]))
  return newestEligible(
    eligible.map((release) => release.version),
    async (version) => dates.get(version),
    now,
  )
}

// body describes how Node is out of date, or returns '' if it isn't.
function body(current, newest) {
  if (!isBehind(current, newest)) return ''
  const [major, newMajor] = [parseVersion(current)[0], parseVersion(newest.version)[0]]
  const lines = [
    `\`web/.nvmrc\` pins Node ${current}, but Node ${newest.version} (released on ${day(newest.date)}) is the newest release that's at least ${minAgeDays} days old, among Node ${major}'s releases and newer major versions' LTS releases.`,
    '',
    `To update, change \`web/.nvmrc\`, and check the [release notes](https://nodejs.org/en/blog/release/v${newest.version}) for anything that affects the front end.`,
    '',
    dockerfileLine('node', newest.version),
  ]
  if (newMajor !== major) {
    lines.push(
      '',
      `It's a new major version, so change \`engines.node\` in \`web/package.json\` to \`>=${newMajor}\` too.`,
    )
  }
  return lines.join('\n')
}

// check returns the issue's body, or '' if Node is up to date. override
// replaces the version read from web/.nvmrc.
async function check({ override, now }) {
  const current = override || readVersion('web/.nvmrc', /^v?(\d+\.\d+\.\d+)\s*$/)
  const found = await newest(await releases(), current, now)
  console.log(`Node: ${current} in use, newest eligible ${found?.version}`)
  return body(current, found)
}

module.exports = { key: 'node', title: 'Node is out of date', check, internals: { candidates, newest, body } }
