# Restructure the Terraform configuration like diary's

For lafeingcrokodil/flashcards#50.

Markers: 🟡 a judgement call where you might prefer the alternative; 🔴
risky or uncertain.

## Design

- **Files, as in diary:**

  | Now | After |
  | --- | --- |
  | `main.tf` (`terraform` block) | `versions.tf` |
  | `google.tf`, the provider in `cloudflare.tf` | `main.tf`: providers, APIs, the project lookup |
  | `artifactregistry.tf` | `artifact_registry.tf` |
  | `githubactions.tf` | `deploy.tf` |
  | `production.tf` | `cloud_run.tf` |
  | the DNS record in `cloudflare.tf` | `dns.tf` (diary has no Cloudflare) |

  Renaming files doesn't touch the state. No `outputs.tf` yet: nothing needs
  an output until lafeingcrokodil/flashcards#52 and #53.
- **The `removed` blocks go**, as the plan for #49 suggested: they did their
  job when #117 was applied.
- **Terraform `>= 1.17`**, as in diary, and 1.17.1 in CI. 1.17.2 is only 11
  days old.
- **Providers pinned exactly, at the newest versions at least 14 days old:**
  `google` 6.46.0 → 8.10.0, `cloudflare` 5.8.2 → 5.28.1. The upgrade guides
  for google 7 and 8 have nothing that affects these resources except the
  next point. `google-beta` isn't added: nothing needs it yet.
- 🟡 **`disable_on_destroy = true` on the APIs.** google 7 dropped that
  default, so without it, removing an API from the list would leave it
  enabled. Keeping `true` keeps what #120 relied on: an API the app stops
  using gets disabled. Diary sets `false`, so that removing an API can't break
  resources that still need it; `plan` shows that case as a destroy, though.
- **The lock file is committed**, with hashes for Linux, macOS (both) and
  Windows, so that `init` on any of them doesn't rewrite it. `infra/.gitignore`
  ignored it (`.terraform*`); its rules move to the root `.gitignore`, as in
  diary, as `.terraform/`.
- **A `terraform` CI job**, as in diary: `fmt -check -recursive`,
  `init -backend=false`, `validate`, with `hashicorp/setup-terraform` v4.0.1
  pinned by SHA like our other actions.
- **Dependabot updates the providers** (`/infra`), as in diary.
- 🟡 **Nothing checks for a new Terraform CLI version**, which Dependabot
  doesn't update. I'd add it to the scope of lafeingcrokodil/flashcards#93,
  rather than to `versions.yml` here.
- **Left out:** diary's other differences (descriptions, variable defaults,
  `depends_on` on the APIs, the Workload Identity condition), which belong to
  #52–#54.

## Risks

- **The provider upgrade replaces or destroys something:** `terraform plan`
  should show no changes, or only in-place updates of provider-computed
  attributes (for example on `cloudflare_dns_record.production`), and in
  particular no replacements or destroys.
- **`apply` can't easily be rolled back:** the state is then written by the
  new providers, and going back to the old ones means editing the state. No
  check catches this; `plan` beforehand is the safeguard.
- **Invalid or unformatted configuration:** the new CI job.
- **The lock file lacks your platform's hash:** `terraform init` then adds it,
  which shows up as a change in `git status`.
- **Dependabot ignores the Terraform entry:** a follow-up check once it's on
  `main`.

## Careful review

- `.github/workflows/ci.yml`'s new job: a new action dependency
  (`hashicorp/setup-terraform`). It runs with the workflow's read-only
  permissions and no secrets.
- The provider upgrade: the `terraform plan` check in "Risks", since it's hard
  to undo.
