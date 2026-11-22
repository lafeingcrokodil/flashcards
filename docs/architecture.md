# Architecture

A web app for learning languages, with flashcards and reading practice. The
goal is to rebuild the features of the [prototype](../prototype/README.md) with
a proper back end and front end, tests and deployment.

Planned work is tracked as issues in the repository's
[milestones](https://github.com/lafeingcrokodil/flashcards/milestones), and
done in small PRs that can each be reviewed comfortably on their own.

- **Back end:** Go, using the standard library's `net/http`. Each area of the
  app (packs, progress, ...) has a storage interface with an in-memory
  implementation (for unit tests) and a Firestore implementation (tested
  against the Firestore emulator).
- **Front end:** SvelteKit + TypeScript, built as a single-page app with
  `adapter-static`, mobile-first. The built assets are embedded into the Go
  binary with `go:embed`, so the whole app is a single container.
- **Languages:** the app doesn't depend on any particular language. Which
  languages are learned, and from which, is part of the data.
- **Auth:** Firebase Authentication with the Google provider. The front end
  obtains an ID token; the Go server verifies it and checks the email against
  an allowlist.
- **Infrastructure:** GCP, in the existing project `lafeingcrokodil-flashcards`,
  managed with Terraform (state in a GCS bucket): Cloud Run, Firestore,
  Artifact Registry and Workload Identity Federation, with DNS records in
  Cloudflare. The project and the state bucket are created by hand. There is
  a single environment, without staging.
