variable "project_id" {
  type        = string
  description = "The GCP project ID"
}

variable "project_name" {
  type        = string
  description = "The GCP project name"
}

variable "billing_account_id" {
  type        = string
  description = "The GCP billing account ID"
}

variable "region" {
  description = "Default region for regional resources"
  type        = string
}

variable "zone" {
  description = "Default zone for zonal resources"
  type        = string
}

variable "terraform_bucket" {
  description = "The name of the Google Cloud Storage bucket for storing the Terraform state"
  type        = string
}

variable "cloudflare_api_token" {
  description = "The Cloudflare API token"
  type        = string
  sensitive   = true
}

variable "cloudflare_zone_id" {
  description = "The Cloudflare zone ID"
  type        = string
}
