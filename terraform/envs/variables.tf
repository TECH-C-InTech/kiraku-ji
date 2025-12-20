variable "project_id" {
  description = "GCP Project ID"
  type        = string
  default     = "kiraku-ji"
}

variable "region" {
  description = "GCP region"
  type        = string
  default     = "asia-northeast1"
}

variable "service_name" {
  description = "Cloud Run service name"
  type        = string
  default     = "kiraku-ji-api"
}

variable "allow_unauthenticated" {
  description = "Allow unauthenticated access to Cloud Run service"
  type        = bool
  default     = true
}

variable "container_image" {
  description = "Container image URL (will be built by CI/CD)"
  type        = string
  default     = "asia-northeast1-docker.pkg.dev/kiraku-ji/kiraku-ji-api/api:latest"
}

variable "github_org" {
  description = "GitHub organization name"
  type        = string
  default     = "TECH-C-InTech"
}

variable "github_repo" {
  description = "GitHub repository name"
  type        = string
  default     = "kiraku-ji"
}