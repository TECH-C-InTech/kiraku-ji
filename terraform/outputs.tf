output "cloud_run_url" {
  description = "URL of the Cloud Run service"
  value       = google_cloud_run_v2_service.api.uri
}

output "artifact_registry_repository" {
  description = "Artifact Registry repository URL"
  value       = google_artifact_registry_repository.api.name
}

output "service_account_email" {
  description = "Service Account email for Cloud Run"
  value       = google_service_account.cloudrun.email
}
