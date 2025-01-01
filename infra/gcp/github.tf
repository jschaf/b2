resource "google_iam_workload_identity_pool" "jsc-iam-pool-github" {
  workload_identity_pool_id = "jsc-iam-pool-github"
  display_name              = "GitHub Workload Identity Pool"
  description               = "A pool to enable GitHub Actions to authenticate with GCP"
}

resource "google_iam_workload_identity_pool_provider" "jsc-iam-pool-github" {
  workload_identity_pool_id          = google_iam_workload_identity_pool.jsc-iam-pool-github.workload_identity_pool_id
  workload_identity_pool_provider_id = "jsc-iam-pool-github"
  display_name                       = "GitHub Provider"
  description                        = "Provider for GitHub Actions"
  attribute_condition = <<EOT
    assertion.repository_owner == "jschaf" &&
    attribute.repository == "jschaf/jsc" &&
    assertion.ref == "refs/heads/main" &&
    assertion.ref_type == "branch"
EOT
  attribute_mapping = {
    "google.subject"       = "assertion.sub"
    "attribute.actor"      = "assertion.actor"
    "attribute.aud"        = "assertion.aud"
    "attribute.repository" = "assertion.repository"
  }
  oidc {
    issuer_uri = "https://token.actions.githubusercontent.com"
  }
}

resource "google_service_account_iam_binding" "github" {
  service_account_id = google_service_account.blog-releaser.id
  role               = "roles/iam.workloadIdentityUser"
  members = [
    "principalSet://iam.googleapis.com/${google_iam_workload_identity_pool.jsc-iam-pool-github.name}/attribute.repository/jschaf/jsc"
  ]
}
