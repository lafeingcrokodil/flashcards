# Remove staging and the unused APIs

For lafeingcrokodil/flashcards#51.

Markers: 🟡 a judgement call where you might prefer the alternative; 🔴
risky or uncertain.

## Design

- **Staging's Terraform resources are deleted from the configuration**, so
  that `apply` destroys them: all of `infra/staging.tf` (service account, its
  IAM members, domain mapping) and `cloudflare_dns_record.staging`. Rejected:
  `removed` blocks with `destroy = false`, which would leave them in GCP and
  Cloudflare unmanaged.
- 🟡 **The staging Cloud Run service is deleted by hand**, with
  `gcloud run services delete staging`, straight after `terraform apply`,
  since Terraform doesn't manage it. Rejected: an `import` block, so that
  Terraform destroys it. That takes two applies (import, then delete) and a
  resource definition written only to be deleted.
- **Sheets, Compute and OS Login are removed from `enable_apis`.**
  `google_project_service` disables an API when it's destroyed, which is the
  point: an enabled API the app doesn't use is attack surface.
- **No other files change.** Nothing outside `infra/` refers to staging, and
  `docs/architecture.md` already describes a single environment.

## Risks

- **`apply` touches production:** `terraform plan` should show 8 destroys
  (the 4 resources in `infra/staging.tf`, the staging DNS record and the 3
  APIs) and nothing else.
- 🔴 **Disabling Compute breaks or deletes something.** Disabling an API with
  live resources can fail or stop them, and Cloud Run services run as the
  Compute default service account unless told otherwise. Before `apply`:
  `gcloud compute instances list` and `gcloud compute disks list` should be
  empty, and production's service account (`gcloud run services describe
  production --format='value(spec.template.spec.serviceAccountName)'`) should
  be `production@…`, not `…-compute@…`. I can't check these from here.
- **OS Login is disabled before Compute** and GCP refuses, because Compute
  depends on it. Terraform destroys `for_each` members in no fixed order, so
  the first `apply` may fail on OS Login. Running it again then succeeds:
  nothing is left half-done.
- **Invalid configuration:** `terraform fmt -check` and `terraform validate`,
  locally.

## Careful review

- 🔴 Removing `compute.googleapis.com` from `infra/google.tf`: the checks in
  "Risks" before `apply`. A disabled API can be re-enabled, but whatever it
  stopped might not come back as it was.
