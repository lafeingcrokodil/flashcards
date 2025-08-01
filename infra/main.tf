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

resource "google_storage_bucket" "terraform_state" {
  name          = var.terraform_bucket
  location      = var.region
  storage_class = "REGIONAL"

  uniform_bucket_level_access = true

  versioning {
    enabled = true
  }
}
