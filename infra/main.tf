terraform {
  required_version = ">= 1.12.2"
  required_providers {
    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = "5.8.2"
    }
    google = {
      source  = "hashicorp/google"
      version = "6.46.0"
    }
  }
  # The state bucket can't be managed by the configuration that stores its
  # state in it, so it's created by hand (see README.md).
  backend "gcs" {
    bucket = "lafeingcrokodil-flashcards-terraform"
  }
}

# The state bucket used to be managed here. This removes it from the state
# without deleting it.
removed {
  from = google_storage_bucket.terraform_state
  lifecycle {
    destroy = false
  }
}
