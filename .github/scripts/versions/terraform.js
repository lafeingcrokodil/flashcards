// Checks Terraform: the version of the terraform job in ci.yml, against the
// releases in HashiCorp's release API.
const { minAgeDays, parseVersion, newestEligible, isBehind, readVersion, fetchText, day } = require('./common.js')

// The API lists releases newest first, by when they were published. 20 (the
// most it returns at once) is a couple of months of releases, so it includes
// one that's old enough unless releases come unusually often.
const api = 'https://api.releases.hashicorp.com/v1/releases/terraform?limit=20'

// releases maps Terraform's recent stable versions (e.g. "1.17.1") to their
// publish dates.
async function releases() {
  const list = JSON.parse(await fetchText(api))
  return new Map(
    list
      .filter((release) => parseVersion(release.version))
      .map((release) => [release.version, new Date(release.timestamp_created)]),
  )
}

// body describes how Terraform is out of date, or returns '' if it isn't.
function body(current, newest) {
  if (!isBehind(current, newest)) return ''
  return [
    `CI uses Terraform ${current}, but ${newest.version} (released on ${day(newest.date)}) is the newest release that's at least ${minAgeDays} days old.`,
    '',
    `To update, change the \`terraform_version\` of the \`terraform\` job in \`.github/workflows/ci.yml\`, and check the [changelog](https://github.com/hashicorp/terraform/releases/tag/v${newest.version}) for anything that affects the configuration in \`infra/\`. If the new version is a new minor version, also consider raising \`required_version\` in \`infra/versions.tf\` and the version in \`infra/README.md\`.`,
  ].join('\n')
}

// check returns the issue's body, or '' if Terraform is up to date. override
// replaces the version read from ci.yml.
async function check({ override, now }) {
  const current = override || readVersion('.github/workflows/ci.yml', /terraform_version: (\S+)/)
  const dates = await releases()
  const newest = await newestEligible([...dates.keys()], async (version) => dates.get(version), now)
  console.log(`Terraform: ${current} in use, newest eligible ${newest?.version}`)
  return body(current, newest)
}

module.exports = { key: 'terraform', title: 'Terraform is out of date', check, internals: { body } }
