# web

The app's front end: a [SvelteKit](https://svelte.dev/docs/kit) app written in
TypeScript and built as a single-page app with `@sveltejs/adapter-static`.

## Development

Requires the Node.js version in `.nvmrc` (or later) and the pnpm version in
the `packageManager` field of `package.json`, e.g. installed with
`npm install --global pnpm@<version>`.

```bash
pnpm install   # install dependencies
pnpm dev       # start a dev server on http://localhost:5173
pnpm check     # type-check
pnpm lint      # check formatting (Prettier) and lint (ESLint)
pnpm format    # fix formatting
pnpm build     # build the app into build/
```

pnpm only installs package versions that were published at least 14 days ago
(`minimumReleaseAge` in `pnpm-workspace.yaml`).
