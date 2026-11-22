// Checks whether the tools in versions/ are behind the newest release that's
// at least 14 days old, and reports each as a GitHub issue with
// report-issue.js: opened while it's out of date, closed once it isn't.
//
// It's meant to be run with actions/github-script, from the repository's
// root.
const report = require('./report-issue.js')

const tools = ['go', 'golangci-lint', 'node', 'pnpm'].map((name) => require(`./versions/${name}.js`))

// main runs the check. overrides maps a tool's key (go, lint, node or pnpm)
// to a version that replaces the one read from the files, so that a manual
// run can test what happens when one is out of date. Each tool is reported
// before the next one is checked, so that one source being unreachable
// doesn't stop the reports before it.
async function main({ github, context, overrides = {}, now = new Date() }) {
  for (const tool of tools) {
    const body = await tool.check({ override: overrides[tool.key], now })
    await report({ github, context, title: tool.title, body })
  }
}

module.exports = main
