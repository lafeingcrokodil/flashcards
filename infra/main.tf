provider "google" {
  project = var.project_id
  region  = var.region
  zone    = var.zone
}

provider "cloudflare" {
  api_token = var.cloudflare_api_token
}

# Used to look up the project number.
data "google_project" "project" {
  project_id = var.project_id
}

resource "google_project_service" "enable_apis" {
  for_each = toset([
    "artifactregistry.googleapis.com",
    "firebaserules.googleapis.com",
    "firestore.googleapis.com",
    "iamcredentials.googleapis.com",
    "run.googleapis.com",
  ])
  service = each.key

  # Disable an API when it's removed from the list, so that APIs the app no
  # longer uses don't stay enabled.
  disable_on_destroy = true
}
