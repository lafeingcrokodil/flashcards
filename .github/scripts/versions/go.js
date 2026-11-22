// Checks Go: the go directive in go.mod, against the releases of the
// golang.org/toolchain module on the Go module proxy.
const { minAgeDays, parseVersion, newestEligible, isBehind, readVersion, fetchText, day } = require('./common.js')

const proxy = 'https://proxy.golang.org'
const toolchainModule = 'golang.org/toolchain'

// releases lists Go's stable releases (e.g. "1.27.4"). The toolchain module
// has a version per release and platform, so one platform's are used.
async function releases() {
  const list = await fetchText(`${proxy}/${toolchainModule}/@v/list`)
  return list
    .split('\n')
    .map((line) => /^v0\.0\.1-go(.+)\.linux-amd64$/.exec(line.trim()))
    .filter((match) => match && parseVersion(match[1]))
    .map((match) => match[1])
}

async function date(version) {
  const info = JSON.parse(await fetchText(`${proxy}/${toolchainModule}/@v/v0.0.1-go${version}.linux-amd64.info`))
  return new Date(info.Time)
}

// body describes how Go is out of date, or returns '' if it isn't.
function body(current, newest) {
  if (!isBehind(current, newest)) return ''
  return [
    `\`go.mod\` requires Go ${current}, but Go ${newest.version} (released on ${day(newest.date)}) is the newest release that's at least ${minAgeDays} days old.`,
    '',
    `To update, run \`go mod edit -go=${newest.version}\`, and check the [release notes](https://go.dev/doc/devel/release) for anything that affects the code.`,
  ].join('\n')
}

// check returns the issue's body, or '' if Go is up to date. override
// replaces the version read from go.mod.
async function check({ override, now }) {
  const current = override || readVersion('go.mod', /^go (\S+)$/m)
  const newest = await newestEligible(await releases(), date, now)
  console.log(`Go: ${current} in use, newest eligible ${newest?.version}`)
  return body(current, newest)
}

module.exports = { key: 'go', title: 'Go is out of date', check, internals: { body } }
