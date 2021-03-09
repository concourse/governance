# TODO: discord roles too
#
# org member = contributor role
# org owner = admin role
# team maps to team in discord

# resource "github_membership" "org" {
#   for_each = local.people

#   username = each.value.github
#   role = each.value.admin ? "admin" : "member"
# }

resource "github_team" "teams" {
  for_each = local.teams

  name = each.value.name
  description = trimspace(join(" ", split("\n", each.value.purpose)))
  privacy = "closed"
  create_default_maintainer = false
}

resource "github_repository" "repos" {
  for_each = local.repos

  name = each.value.name
  description = trimspace(join(" ", split("\n", each.value.description)))

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

  has_issues = try(each.value.has_issues, true)
  has_projects = try(each.value.has_projects, false)
  has_wiki = try(each.value.has_wiki, false)

  # this was deprecated in 2013 but still defaults to true?
  has_downloads = false

  archive_on_destroy = true
}

resource "github_team_membership" "members" {
  for_each = {
    for membership in local.team_memberships :
      "${membership.team_name}:${membership.username}" => membership
  }

  team_id = github_team.teams[each.value.team_name].id
  username = each.value.username
  role = each.value.role
}

resource "github_team_repository" "repos" {
  for_each = {
    for ownership in local.team_repos :
      "${ownership.team_name}:${ownership.repository}" => ownership
  }

  team_id = github_team.teams[each.value.team_name].id
  repository = each.value.repository
  permission = each.value.permission
}
