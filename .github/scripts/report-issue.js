// Reports the result of a check as a GitHub issue, so that a problem found by
// a scheduled run doesn't go unnoticed. There's at most one open issue per
// title, labelled "automated":
//
// - If the check found a problem (a non-empty body), the issue is opened, or
//   its description replaced if the details changed.
// - If it didn't (an empty body), an open issue is closed.
//
// It's meant to be run with actions/github-script, which provides github (an
// authenticated API client) and context.
const label = 'automated'

module.exports = async ({ github, context, title, body }) => {
  const { owner, repo } = context.repo

  // Issues with other titles or without the label, such as ones opened by
  // hand, are never touched.
  const issues = await github.paginate(github.rest.issues.listForRepo, {
    owner,
    repo,
    state: 'open',
    labels: label,
    per_page: 100,
  })
  // The issues API also lists pull requests.
  const existing = issues.find(
    (issue) => !issue.pull_request && issue.title === title,
  )
  const found = body.trim() !== ''

  if (found && !existing) {
    const { data } = await github.rest.issues.create({
      owner,
      repo,
      title,
      body,
      labels: [label],
      assignees: [owner],
    })
    console.log(`Opened #${data.number}.`)
  } else if (found && existing.body !== body) {
    await github.rest.issues.update({
      owner,
      repo,
      issue_number: existing.number,
      body,
    })
    console.log(`Updated #${existing.number}.`)
  } else if (found) {
    console.log(`#${existing.number} is already up to date.`)
  } else if (existing) {
    await github.rest.issues.createComment({
      owner,
      repo,
      issue_number: existing.number,
      body: 'The check passes again, so this issue was closed automatically.',
    })
    await github.rest.issues.update({
      owner,
      repo,
      issue_number: existing.number,
      state: 'closed',
      state_reason: 'completed',
    })
    console.log(`Closed #${existing.number}.`)
  } else {
    console.log('Nothing to report.')
  }
}
