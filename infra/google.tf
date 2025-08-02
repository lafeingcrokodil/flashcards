provider "google" {
  project = var.project_id
  region  = var.region
  zone    = var.zone
}

resource "google_project" "project" {
  name            = var.project_name
  project_id      = var.project_id
  billing_account = var.billing_account_id

  auto_create_network = false
}

resource "google_project_service" "enable_apis" {
  for_each = toset([
    "artifactregistry.googleapis.com",
    "compute.googleapis.com",
    "firebaserules.googleapis.com",
    "firestore.googleapis.com",
    "iamcredentials.googleapis.com",
    "oslogin.googleapis.com",
    "run.googleapis.com",
    "sheets.googleapis.com",
  ])
  service = each.key
}
