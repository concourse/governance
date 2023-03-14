terraform {
  required_providers {
    github = {
      source = "integrations/github"
    }

    mailgun = {
      source = "wgebis/mailgun"
    }
  }
}

provider "github" {
  token = var.github_token
  owner = "concourse"
}

provider "mailgun" {
  api_key = var.mailgun_api_key
}
