// Checks golangci-lint: the version of the lint job in ci.yml, and the one in
// the README, against the releases of its module on the Go module proxy.
const { minAgeDays, parseVersion, newestEligible, isBehind, readVersion, fetchText, day } = require('./common.js')

const proxy = 'https://proxy.golang.org'
const lintModule = 'github.com/golangci/golangci-lint/v2'

async function releases() {
  const list = await fetchText(`${proxy}/${lintModule}/@v/list`)
  return list
    .split('\n')
    .map((line) => line.trim())
    .filter((version) => parseVersion(version))
}

async function date(version) {
  const info = JSON.parse(await fetchText(`${proxy}/${lintModule}/@v/${version}.info`))
  return new Date(info.Time)
}

// body describes how golangci-lint is out of date, or how the README
// disagrees with CI's version (ci, which current only differs from when it's
// overridden), or returns '' if neither.
function body(current, ci, readme, newest) {
  const lines = []
  if (isBehind(current, newest)) {
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

// check returns the issue's body, or '' if golangci-lint is up to date.
// override replaces the version read from ci.yml.
async function check({ override, now }) {
  const ci = readVersion('.github/workflows/ci.yml', /golangci-lint-action@.*\n(?:.*\n)*?\s*version: (\S+)/)
  const current = override || ci
  const readme = readVersion('README.md', /golangci-lint\]\(.*\) (v\S+)/)
  const newest = await newestEligible(await releases(), date, now)
  console.log(`golangci-lint: ${current} in use (README: ${readme}), newest eligible ${newest?.version}`)
  return body(current, ci, readme, newest)
}

module.exports = { key: 'lint', title: 'golangci-lint is out of date', check, internals: { body } }
