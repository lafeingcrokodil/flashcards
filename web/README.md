# web

The app's front end: a [SvelteKit](https://svelte.dev/docs/kit) app written in
TypeScript and built as a single-page app with `@sveltejs/adapter-static`.

## Development

Requires the Node.js version in `.nvmrc` (or later) and the pnpm version in
the `packageManager` field of `package.json`, e.g. installed with
`npm install --global pnpm@<version>`.

```bash
pnpm install   # install dependencies
pnpm dev       # start a dev server on http://localhost:5173 (see below)
pnpm check     # type-check
pnpm lint      # check formatting (Prettier) and lint (ESLint)
pnpm format    # fix formatting
pnpm build     # build the app into build/
```

In development, the dev server forwards requests under `/api` to the Go server
on http://localhost:8080 (see the [root README](../README.md)). In production,
the Go server serves the built app itself.

pnpm only installs package versions that were published at least 14 days ago
(`minimumReleaseAge` in `pnpm-workspace.yaml`).
