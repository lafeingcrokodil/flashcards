resource "google_service_account" "github_actions" {
  account_id = "github-actions"
}

resource "google_project_iam_member" "github_actions" {
  project = google_project.project.id
  role    = "roles/datastore.user"
  member  = google_service_account.github_actions.member
}

resource "google_iam_workload_identity_pool" "github" {
  workload_identity_pool_id = "github"
}

resource "google_iam_workload_identity_pool_provider" "github" {
  workload_identity_pool_id          = google_iam_workload_identity_pool.github.workload_identity_pool_id
  workload_identity_pool_provider_id = "github"
  attribute_condition                = <<EOT
    attribute.repository_owner == "lafeingcrokodil"
EOT
  attribute_mapping = {
    "google.subject"             = "assertion.sub"
    "attribute.repository_owner" = "assertion.repository_owner"
    "attribute.repository"       = "assertion.repository"
  }
  oidc {
    issuer_uri = "https://token.actions.githubusercontent.com"
  }
}

resource "google_service_account_iam_member" "github_workload_identity" {
  service_account_id = google_service_account.github_actions.id
  role               = "roles/iam.workloadIdentityUser"
  member             = "principalSet://iam.googleapis.com/projects/${google_project.project.number}/locations/global/workloadIdentityPools/github/attribute.repository/lafeingcrokodil/flashcards"
}
