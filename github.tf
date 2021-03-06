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
