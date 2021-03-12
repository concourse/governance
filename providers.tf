terraform {
  required_providers {
    github = {
      source  = "integrations/github"
      version = "4.5.0"
    }
  }
}

provider "github" {
  token        = var.github_token
  organization = "concourse"
}
