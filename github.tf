resource "github_membership" "contributors" {
  for_each = local.contributors

  username = each.value.github
  role     = "member"
}

resource "github_team" "teams" {
  for_each = local.teams

  name        = each.value.name
  description = trimspace(join(" ", split("\n", each.value.purpose)))
  privacy     = "closed"

  create_default_maintainer = false

  # TODO: remove once we remove the old team hierarchy
  lifecycle {
    ignore_changes = [parent_team_id, privacy]
  }
}

resource "github_repository" "repos" {
  for_each = local.repos

  name        = each.value.name
  description = trimspace(join(" ", split("\n", each.value.description)))

  visibility = try(each.value.private, false) ? "private" : "public"

  # TODO: this has caused errors with a newly created repo before. maybe an API
  # race condition?
  #
  #   Error: PUT https://api.github.com/repos/concourse/foo/topics: 404 Not Found []
  #
  # it's fixable by untainting the resource to prevent it from deleting the
  # repo and applying again:
  #
  #   terraform untaint 'github_repository.repos["foo"]'
  #   terraform apply
  topics = try(each.value.topics, [])

  homepage_url = try(each.value.homepage_url, null)
  has_issues   = try(each.value.has_issues, false)
  has_projects = try(each.value.has_projects, false)
  has_wiki     = try(each.value.has_wiki, false)

  # this was deprecated in 2013 but still defaults to true?
  has_downloads = false

  # safer sane default; repo can be manually destroyed if desired
  archive_on_destroy = true

  # sane defaults
  vulnerability_alerts   = true
  delete_branch_on_merge = true

  dynamic "pages" {
    for_each = try([each.value.pages], [])

    content {
      cname = pages.value.cname
      source {
        branch = pages.value.branch
        path   = try(pages.value.path, null)
      }
    }
  }
}

resource "github_branch_protection" "branch_protections" {
  for_each = {
    for protection in local.repo_branch_protections :
    "${protection.repository_name}:${protection.pattern}" => protection
  }

  repository_id = github_repository.repos[each.value.repository_name].node_id
  pattern       = each.value.pattern

  allows_deletions = each.value.allows_deletions

  required_status_checks {
    contexts = each.value.required_checks
    strict   = each.value.strict_checks
  }

  dynamic "required_pull_request_reviews" {
    for_each = each.value.required_reviews == 0 ? [] : [each.value]

    content {
      required_approving_review_count = each.value.required_reviews
      dismiss_stale_reviews           = each.value.dismiss_stale_reviews
      require_code_owner_reviews      = each.value.require_code_owner_reviews
    }
  }

  # force pushing is generally not a great idea, so let's set this to false
  # until someone has a good reason to make it configurable
  allows_force_pushes = false

  # there are no repository admins to inconvenience in this model, so we might
  # as well play it safe
  enforce_admins = true
}

resource "github_team_membership" "members" {
  for_each = {
    for membership in local.team_memberships :
    "${membership.team_name}:${membership.username}" => membership
  }

  team_id  = github_team.teams[each.value.team_name].id
  username = each.value.username
  role     = each.value.role
}

resource "github_team_repository" "repos" {
  for_each = {
    for ownership in local.team_repos :
    "${ownership.team_name}:${ownership.repository}" => ownership
  }

  team_id    = github_team.teams[each.value.team_name].id
  repository = github_repository.repos[each.value.repository].name
  permission = each.value.permission
}

resource "github_repository_collaborator" "collaborators" {
  for_each = {
    for c in local.repo_collaborators :
    "${c.repository}:${c.username}" => c
  }

  repository = github_repository.repos[each.value.repository].name
  username   = each.value.username
  permission = each.value.permission
}
