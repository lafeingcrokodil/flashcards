resource "google_cloud_run_v2_service" "staging" {
  name     = "staging"
  location = var.region

  template {
    containers {
      image = var.image_url
    }
  }
}

resource "google_cloud_run_v2_service_iam_member" "staging_invoker" {
  name     = google_cloud_run_v2_service.staging.name
  location = google_cloud_run_v2_service.staging.location
  role     = "roles/run.invoker"
  for_each = toset(var.staging_invokers)
  member   = each.value
}
