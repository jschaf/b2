resource "google_artifact_registry_repository" "jsc-art-uswe2-docker" {
  location               = "us-west2"
  repository_id          = "jsc-art-uswe2-docker"
  description            = "Docker registry for prod images"
  format                 = "DOCKER"
}
