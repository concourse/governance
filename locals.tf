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
        role      = "member"
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

  repo_branch_protections = flatten([
    for repo in local.repos : [
      for protection in try(repo.branch_protection, []) : {
        repository_name = repo.name
        pattern         = protection.pattern

        allows_deletions = try(protection.allows_deletions, false)

        required_checks = try(protection.required_checks, [])
        strict_checks   = try(protection.strict_checks, false)

        required_reviews           = try(protection.required_reviews, 0)
        dismiss_stale_reviews      = try(protection.dismiss_stale_reviews, false)
        require_code_owner_reviews = try(protection.require_code_owner_reviews, false)
      }
    ]
  ])

  repo_issue_labels = flatten([
    for repo in local.repos : [
      for label in try(repo.labels, []) : {
        repository_name = repo.name

        name  = label.name
        color = label.color
      }
    ]
  ])
}
