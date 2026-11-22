variable "project_id" {
  type        = string
  description = "The GCP project ID"
}

variable "region" {
  description = "Default region for regional resources"
  type        = string
}

variable "zone" {
  description = "Default zone for zonal resources"
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
