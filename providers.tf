terraform {
  required_providers {
    github = {
      source  = "integrations/github"
      version = "4.6.0"
    }

    mailgun = {
      source  = "wgebis/mailgun"
      version = "0.6.1"
    }
  }
}

provider "github" {
  token = var.github_token
  organization = "concourse"
}

provider "mailgun" {
  api_key = var.mailgun_api_key
}
