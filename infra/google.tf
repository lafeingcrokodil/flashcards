provider "google" {
  project = var.project_id
  region  = var.region
  zone    = var.zone
}

# The project has to exist before Terraform can create anything in it, so
# it's created by hand (see README.md). It used to be managed here: this
# removes it from the state without deleting it.
removed {
  from = google_project.project
  lifecycle {
    destroy = false
  }
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
}
