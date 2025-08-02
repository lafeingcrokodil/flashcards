resource "google_service_account" "production" {
  account_id = "production"
}

resource "google_project_iam_member" "production" {
  project = google_project.project.id
  for_each = toset([
    "roles/datastore.user",
  ])
  role   = each.value
  member = google_service_account.production.member
}

resource "google_service_account_iam_member" "production_users" {
  service_account_id = google_service_account.production.id
  role               = "roles/iam.serviceAccountUser"
  member             = google_service_account.github_actions.member
}
