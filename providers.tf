terraform {
  required_providers {
    github = {
      source = "integrations/github"
      version = "4.5.0"
    }

    discord = {
      source = "aequasi/discord"
      version = "0.0.4"
    }
  }
}

provider "github" {
  token = var.github_token
  organization = "concourse"
}

provider "discord" {
  token = var.discord_token
}
