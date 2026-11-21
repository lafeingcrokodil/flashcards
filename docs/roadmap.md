# Roadmap

A web app for learning languages, with flashcards and reading practice. The
goal is to rebuild the features of the [prototype](../prototype/README.md) with
a proper back end and front end, tests and deployment.

The app is built incrementally: each milestone is split into small PRs that
can each be reviewed comfortably on their own.

## Architecture

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
  Cloudflare. There is a single environment, without staging.

## Milestones

### Before M0 – Development workflow

- [x] Extend `AGENTS.md` to describe a development workflow that includes
      planning features before they're implemented

### M0 – Walking skeleton

- [x] Go module, `/health` endpoint, test, Makefile
- [ ] GitHub Actions CI (`go vet`, `go test`, `golangci-lint` with all linters
      enabled), with the `main` ruleset requiring its jobs
- [ ] Dependabot version updates, with a 14-day cooldown
- [ ] SvelteKit + TypeScript app scaffold, with linting in CI
- [ ] Serve the front end from the Go binary, local dev setup
- [ ] Dockerfile

### M1 – Deployed to GCP

- [ ] Stop managing the project and the state bucket with Terraform, and name
      the bucket in the backend configuration
- [ ] Restructure the Terraform configuration like diary's (file layout,
      minimum Terraform version, current providers, committed lock file), with
      `terraform fmt` and `terraform validate` in CI
- [ ] Remove what the app no longer needs: staging (service account, IAM,
      domain mapping, DNS record and Cloud Run service) and unused APIs
      (Sheets, Compute, OS Login)
- [ ] Cloud Run managed by Terraform, and Artifact Registry with a cleanup
      policy
- [ ] Continuous deployment from `main` via Workload Identity Federation
- [ ] Firestore database with delete protection (and abandoned rather than
      deleted by Terraform), and a backup schedule, possibly with
      point-in-time recovery
- [ ] Budget alert

### M2 – Authentication

- [ ] Firebase project, and serving the app on
      `flashcards.lafeingcrokodil.com`
- [ ] Firebase Authentication with the Google provider
- [ ] Front-end sign-in flow
- [ ] Go middleware verifying ID tokens against an email allowlist

### M3 – Packs

- [ ] Domain types and storage interface for packs of words (each word with
      its translation, an example sentence and its translation, an optional
      note and optional alternative answers)
- [ ] CRUD API and Firestore implementation
- [ ] Front end for managing packs and their words

### M4 – Cards

- [ ] Leitner scheduling: a right answer moves a word up a box (due in 1, 2,
      4, 8, 16 or 32 days), a wrong one sends it back to the first box
- [ ] Storing each word's progress
- [ ] Review sessions (due words first, then a chosen number of new ones), with
      wrong answers repeated at the end
- [ ] Filtering by pack, and progress per pack

### M5 – Type

- [ ] Typed recall in the other direction, sharing the scheduling with Cards
- [ ] Alternative answers, an optional hint and an "I was right" override

### M6 – Read

- [ ] Passages with lines, translations, questions and a glossary
- [ ] Multiple-choice questions
- [ ] Tapping a word to see its glossary entry, and saving it to a pack
- [ ] Sending a word back to the first box ("Review again")

### M7 – Feedback, lookups and speech

Features that the prototype implements with Claude. They might use an AI
model in the app too, but don't have to.

- [ ] Writing sentences with a word, with feedback and a corrected sentence
- [ ] Translating passage lines, with feedback
- [ ] Looking up words that aren't in a passage's glossary
- [ ] Hearing words read out (e.g. with the browser's speech synthesis)

### Later ideas

- A staging environment, in a separate GCP project
- Installable PWA

## Open questions

To be discussed when we reach the relevant milestone:

- Is the app served on its custom domain through Firebase Hosting (as in
  diary) or through Cloud Run domain mapping?
- Where do packs and passages come from (entered by hand, imported from
  files, generated)?
