resource "google_artifact_registry_repository" "main" {
  repository_id = "main"
  format        = "DOCKER"
  location      = var.region
}
