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

resource "google_project_service" "secretmanager" {
  service            = "secretmanager.googleapis.com"
  disable_on_destroy = false
}

resource "google_project_service" "firestore" {
  service            = "firestore.googleapis.com"
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
  account_id   = "cloudrun"
  display_name = "Service Account for Cloud Run (${var.service_name})"
}

# Grant Cloud Run service account access to secrets
resource "google_secret_manager_secret_iam_member" "cloudrun_openai_api_key" {
  secret_id = "OPENAI_API_KEY"
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.cloudrun.email}"

  depends_on = [google_project_service.secretmanager]
}

resource "google_secret_manager_secret_iam_member" "cloudrun_cors_allow_origins" {
  secret_id = "CORS_ALLOW_ORIGINS"
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.cloudrun.email}"

  depends_on = [google_project_service.secretmanager]
}

# Cloud Run service
resource "google_cloud_run_v2_service" "api" {
  name                = var.service_name
  location            = var.region
  deletion_protection = false # Allow Terraform to delete/recreate service

  template {
    service_account = google_service_account.cloudrun.email

    # Resource limits
    scaling {
      min_instance_count = 0
      max_instance_count = 10
    }

    containers {
      image = var.container_image

      ports {
        container_port = 8080
      }

      resources {
        limits = {
          cpu    = "1"
          memory = "512Mi"
        }
      }

      # Environment variables
      env {
        name  = "GOOGLE_CLOUD_PROJECT"
        value = var.project_id
      }

      env {
        name  = "LLM_PROVIDER"
        value = "openai"
      }

      env {
        name = "OPENAI_API_KEY"
        value_source {
          secret_key_ref {
            secret  = "OPENAI_API_KEY"
            version = "latest"
          }
        }
      }

      env {
        name = "CORS_ALLOW_ORIGINS"
        value_source {
          secret_key_ref {
            secret  = "CORS_ALLOW_ORIGINS"
            version = "latest"
          }
        }
      }
    }

    timeout         = "60s"
    max_instance_request_concurrency = 80
  }

  lifecycle {
    ignore_changes = [
      # デプロイは GitHub Actions に寄せる（Terraform はインフラ設定を正とする）
      template[0].containers[0].image,
    ]
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
