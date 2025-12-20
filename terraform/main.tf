# Enable required APIs
resource "google_project_service" "run" {
  service            = "run.googleapis.com"
  disable_on_destroy = false
}

resource "google_project_service" "artifactregistry" {
  service            = "artifactregistry.googleapis.com"
  disable_on_destroy = false
}

resource "google_project_service" "iam" {
  service            = "iam.googleapis.com"
  disable_on_destroy = false
}

# Artifact Registry repository
resource "google_artifact_registry_repository" "api" {
  location      = var.region
  repository_id = var.service_name
  description   = "Docker repository for kiraku-ji API"
  format        = "DOCKER"

  depends_on = [google_project_service.artifactregistry]
}

# Service Account for Cloud Run
resource "google_service_account" "cloudrun" {
  account_id   = "${var.service_name}-sa"
  display_name = "Service Account for Cloud Run (${var.service_name})"
}

# Cloud Run service
resource "google_cloud_run_v2_service" "api" {
  name                = var.service_name
  location            = var.region
  deletion_protection = false # Allow Terraform to delete/recreate service

  template {
    service_account = google_service_account.cloudrun.email

    containers {
      image = var.container_image

      ports {
        container_port = 8080
      }

      # Environment variables can be added here
      # env {
      #   name  = "CORS_ALLOW_ORIGINS"
      #   value = "*"
      # }
    }
  }

  depends_on = [google_project_service.run]
}

# Allow unauthenticated access (public API)
resource "google_cloud_run_v2_service_iam_member" "public" {
  count = var.allow_unauthenticated ? 1 : 0

  location = google_cloud_run_v2_service.api.location
  name     = google_cloud_run_v2_service.api.name
  role     = "roles/run.invoker"
  member   = "allUsers"
}
