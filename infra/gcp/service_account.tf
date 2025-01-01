resource "google_service_account" "blog-releaser" {
  account_id   = "blog-releaser"
  display_name = "blog-releaser"
  description  = "Release the joe.schafer.dev blog"
}
