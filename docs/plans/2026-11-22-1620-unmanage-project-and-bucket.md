# Stop managing the project and the state bucket with Terraform

For lafeingcrokodil/flashcards#49.

Markers: 🟡 a judgement call where you might prefer the alternative; 🔴
risky or uncertain.

## Goal

Take the GCP project (`google_project.project`) and the bucket for
Terraform's state (`google_storage_bucket.terraform_state`) out of the
Terraform configuration, without deleting them, and name the bucket in the
`gcs` backend block instead of passing it to `terraform init`.

The bucket can't be managed by the configuration that keeps its state in it:
deleting or replacing it would take the state with it, and it has to exist
before the first `terraform init`. The project is similar: it has to exist
(with billing) before Terraform can create anything in it, and destroying it
would delete everything else. diary creates both by hand, as described in its
`infra/README.md`.

Left for later:

- The rest of the restructuring like diary's (file layout, `versions.tf`,
  current providers, a committed lock file, `fmt` and `validate` in CI):
  lafeingcrokodil/flashcards#50.
- Removing staging and the unused APIs: lafeingcrokodil/flashcards#51.

## Design

### `removed` blocks, rather than `terraform state rm`

Both resources are replaced in the configuration by `removed` blocks with
`lifecycle { destroy = false }` (Terraform 1.7 and later; the configuration
requires 1.12.2):

```hcl
removed {
  from = google_project.project
  lifecycle {
    destroy = false
  }
}
```

`terraform plan` then shows each of them as "will no longer be managed by
Terraform", and `terraform apply` drops them from the state and leaves them in
GCP. The change is in the PR, so it's reviewed, and `plan` shows it before
anything happens.

- Rejected: `terraform state rm` by hand. It does the same, but it isn't in
  the PR, and if it's forgotten (or done after deleting the resources from the
  configuration), `plan` proposes to destroy the project and the bucket.

Two more guards, independent of the `removed` blocks: the provider's
`deletion_policy` for `google_project` defaults to `PREVENT` (since provider
6.0; the configuration uses 6.46.0), so destroying the project fails, and the
bucket has no `force_destroy`, so destroying it fails while it holds the
state.

🟡 The `removed` blocks stay in the configuration after `terraform apply`:
they're harmless once applied, and lafeingcrokodil/flashcards#50, which
reorganises the files anyway, can delete them.

### References to the project

Three resources refer to `google_project.project`:

- The `google_project_iam_member` resources (`github_actions`, `staging`,
  `production`) use its `id` (`projects/<project_id>`) as `project`. They get
  `var.project_id` instead. `project` forces a new resource when it changes,
  but the provider treats `projects/x` and `x` as the same
  (`CompareProjectName`), so `plan` shouldn't show a change for them.
- `google_service_account_iam_member.github_workload_identity` uses its
  `number` in the principal set. It gets the number from a
  `data "google_project" "project"` instead, as diary's `budget.tf` does. It
  reads the project through the Cloud Resource Manager API, as the
  `google_project` resource did, so no API needs enabling.

### Variables

`project_name` and `billing_account_id` are only used by the project, and
`terraform_bucket` only by the bucket, so they're removed. A
`terraform.tfvars` that still sets them gets a warning ("Value for undeclared
variable"), not an error, and the README says to delete them.

### The bucket's name in the backend block

```hcl
backend "gcs" {
  bucket = "lafeingcrokodil-flashcards-terraform"
}
```

A backend block can't use variables, so the name is written out. It's the
`terraform_bucket` value from your `terraform.tfvars`, which isn't in the
repository, so the name is as you gave it to me rather than checked.

If the name is the same as the one passed to `terraform init` so far,
Terraform doesn't ask to initialise again: I tried this with Terraform 1.12.2
and a `local` backend (a partial configuration with `-backend-config`, then
the same value in the file). If it's different, `terraform plan` stops with
"Backend initialization required", and `terraform init` stops with "Backend
configuration changed" until it's given `-migrate-state` or `-reconfigure`
(tried the same way). So a wrong name fails loudly rather than silently
starting with an empty state.

### Documentation

- `infra/README.md`: like diary's, a "One-off setup" section with the
  `gcloud` commands that create the project (and link billing) and the bucket
  (versioned, uniform bucket-level access, public access prevention), which
  only need running for a new project. Then `terraform init` without
  `-backend-config`, and which variables `terraform.tfvars` still needs. The
  bucket's settings in the commands match the ones in the current
  configuration (regional, in `var.region`, versioned, uniform bucket-level
  access) plus public access prevention, as in diary.
- `docs/architecture.md`: the infrastructure bullet says that the project and
  the state bucket are created by hand.

## Risks

- **`terraform apply` destroys the project or the bucket:** the `removed`
  blocks, and the plan saying "will no longer be managed" rather than
  "destroy". Behind them, the project's `PREVENT` deletion policy and the
  non-empty bucket make a destroy fail. `terraform plan` is the check: it
  should show exactly two resources no longer managed, and no destroys.
- **The IAM members are replaced** (briefly removing a role) because their
  `project` changes from `projects/<id>` to `<id>`: `terraform plan` should
  show no change for them (the provider's `CompareProjectName`). If it does,
  it's harmless (the members are re-added in the same apply), but I'd rather
  you stop and tell me.
- **The workload identity member changes** (if the data source's number
  differs from the resource's): `terraform plan` shows no change for it.
- **The backend names the wrong bucket:** see above; `plan` and `init` stop
  with an error.
- **The configuration is invalid:** CI doesn't check Terraform yet
  (lafeingcrokodil/flashcards#50), so I'll run `terraform fmt -check` and
  `terraform validate` locally with Terraform 1.12.2 and the configuration's
  provider versions.
- **The README's `gcloud` commands are wrong:** not covered by a check, and
  they're only run for a new project. A mistake there is easy to fix and does
  no harm until then.

## Careful review

None. The parts that come closest pass: a wrong bucket name in the backend
block stops `plan` and `init` with an error (see above), a wrong `removed`
block or reference shows up in `terraform plan` (which you run before
`apply`) as a destroy or a replacement, and behind that the deletion policy
and the non-empty bucket stop a destroy; the README and `docs/architecture.md`
are easy to fix.
