// Checks pnpm: the packageManager field in web/package.json, against the
// versions in the npm registry.
const { minAgeDays, parseVersion, compareVersions, newestEligible, isBehind, readVersion, fetchText, day } = require('./common.js')

const registry = 'https://registry.npmjs.org/pnpm'

// releases returns pnpm's stable versions that are no newer than its
// "latest" dist-tag, since newer ones are only tagged "next", and the time
// field of its registry document, which maps each version to its publish
// date.
async function releases() {
  const doc = JSON.parse(await fetchText(registry))
  const latest = doc['dist-tags'].latest
  const versions = Object.keys(doc.time).filter(
    (version) => parseVersion(version) && compareVersions(version, latest) <= 0,
  )
  return { versions, time: doc.time }
}

// body describes how pnpm is out of date, or returns '' if it isn't.
function body(current, newest) {
  if (!isBehind(current, newest)) return ''
  return [
    `\`packageManager\` in \`web/package.json\` says pnpm ${current}, but pnpm ${newest.version} (released on ${day(newest.date)}) is the newest release that's at least ${minAgeDays} days old and no newer than the \`latest\` tag.`,
    '',
    `To update, change \`packageManager\` in \`web/package.json\`, and check the [release notes](https://github.com/pnpm/pnpm/releases/tag/v${newest.version}) for anything that affects the front end.`,
  ].join('\n')
}

// check returns the issue's body, or '' if pnpm is up to date. override
// replaces the version read from web/package.json.
async function check({ override, now }) {
  const current =
    override || readVersion('web/package.json', /"packageManager": "pnpm@(\d+\.\d+\.\d+)(?:\+[^"]*)?"/)
  const { versions, time } = await releases()
  const newest = await newestEligible(versions, async (version) => new Date(time[version]), now)
  console.log(`pnpm: ${current} in use, newest eligible ${newest?.version}`)
  return body(current, newest)
}

module.exports = { key: 'pnpm', title: 'pnpm is out of date', check, internals: { body } }
