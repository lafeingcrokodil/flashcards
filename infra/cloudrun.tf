resource "google_cloud_run_v2_service_iam_member" "staging_invoker" {
  name     = "staging"
  location = var.region
  role     = "roles/run.invoker"
  for_each = toset(var.staging_invokers)
  member   = each.value
}
