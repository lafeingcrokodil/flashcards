# Stop managing the project and the state bucket with Terraform

For lafeingcrokodil/flashcards#49.

Markers: 🟡 a judgement call where you might prefer the alternative; 🔴
risky or uncertain.

## Goal

Take the GCP project and the state bucket out of the Terraform configuration
without deleting them, and name the bucket in the `gcs` backend block
(`lafeingcrokodil-flashcards-terraform`) instead of passing it to
`terraform init`. The README describes creating both by hand, as diary's does.

Left for later: the rest of the restructuring (lafeingcrokodil/flashcards#50)
and removing staging and unused APIs (lafeingcrokodil/flashcards#51).

## Design

- **`removed` blocks with `destroy = false`**, rather than
  `terraform state rm` by hand. They're reviewed in the PR, and can't be
  forgotten: deleting the resources without them would make `plan` propose
  destroying the project and the bucket.
- 🟡 **The `removed` blocks stay after `apply`.** They're harmless then, and
  lafeingcrokodil/flashcards#50 reorganises the files anyway, so it can delete
  them.
- **The project's ID and number** come from `var.project_id` and a
  `data "google_project"`, as in diary.
- **The unused variables** (`project_name`, `billing_account_id`,
  `terraform_bucket`) are removed. A `terraform.tfvars` that still sets them
  only gets a warning.

## Risks

- **`apply` destroys the project or the bucket:** `terraform plan` should show
  exactly two resources "no longer managed by Terraform", and nothing else:
  no destroys, and no replaced IAM members. Behind that, the project's
  default `deletion_policy` (`PREVENT`) and the non-empty bucket make a
  destroy fail.
- **The backend names the wrong bucket:** `plan` and `init` stop with an error
  (checked locally), rather than starting with an empty state.
- **The configuration is invalid:** `terraform fmt -check` and
  `terraform validate`, run locally, since CI doesn't check Terraform yet.

## Careful review

None: each part is covered by `terraform plan`, or (for the README) is easy
to fix.
