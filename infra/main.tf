terraform {
  required_version = "v1.12.2"
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "6.46.0"
    }
  }
  backend "gcs" {
    bucket = "" # specified separately, e.g. via `terraform init` command line option
  }
}

resource "google_storage_bucket" "terraform_state" {
  name          = var.terraform_bucket
  location      = var.region
  storage_class = "REGIONAL"

  uniform_bucket_level_access = true

  versioning {
    enabled = true
  }
}
