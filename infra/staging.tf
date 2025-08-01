resource "google_service_account" "staging" {
  account_id = "staging"
}

resource "google_project_iam_member" "staging" {
  project = google_project.project.id
  for_each = toset([
    "roles/datastore.user",
  ])
  role   = each.value
  member = google_service_account.staging.member
}

resource "google_service_account_iam_member" "staging_users" {
  service_account_id = google_service_account.staging.id
  role               = "roles/iam.serviceAccountUser"
  member             = google_service_account.github_actions.member
}

resource "google_cloud_run_v2_service_iam_member" "staging_invoker" {
  name     = "staging"
  location = var.region
  role     = "roles/run.invoker"
  for_each = toset(var.staging_invokers)
  member   = each.value
}
