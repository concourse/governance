locals {
  contributors = {
    for f in fileset(path.module, "contributors/*.yml") :
    trimsuffix(basename(f), ".yml") => yamldecode(file(f))
  }

  teams = {
    for f in fileset(path.module, "teams/*.yml") :
    trimsuffix(basename(f), ".yml") => yamldecode(file(f))
  }

  repos = {
    for f in fileset(path.module, "repos/*.yml") :
    trimsuffix(basename(f), ".yml") => yamldecode(file(f))
  }

  team_memberships = flatten([
    for team in local.teams : [
      for person in team.members : {
        team_name = team.name
        username  = local.contributors[person].github
        role      = try(local.contributors[person].admin, false) ? "maintainer" : "member"
      }
    ]
  ])

  team_repos = flatten([
    for team in local.teams : [
      for repo in try(team.repos, []) : {
        team_name  = team.name
        repository = repo
        permission = "maintain"
      }
    ]
  ])

  repo_collaborators = flatten([
    for contributor in local.contributors : [
      for repo, permission in try(contributor.repos, {}) : {
        repository = repo
        username   = contributor.github
        permission = permission
      }
    ]
  ])
}
