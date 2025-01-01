terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 6.14.1"
    }
  }
  required_version = "~> 1.10.2"
}

provider "google" {
  // project number 191693554478
  project = "jschaf"
  region  = "us-west2"
  zone    = "us-west2-b"
}
