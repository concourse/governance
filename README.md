# Concourse Governance

This document outlines a set of policies in order to provide a level playing
field and open process for contributors to join the Concourse project.

In addition to this document, this repository contains live configuration for
the state of the Concourse GitHub organization. All configuration is
automatically applied via [Terraform][terraform] and synchronized daily to
prevent drift.

[terraform]: https://www.terraform.io/

## Governance Model

Individual contributors to the Concourse project can become members of
**teams**, each with a stated purpose, a clear set of responsibilities, and a
list of repos that they maintain.

Teams collaborate through discussions on GitHub and propose changes through
pull requests that may cross team boundaries.

Ideally, teams should be split along boundaries that enhance the focus given to
different facets of the Concourse project. Repositories should typically belong
to a single team in order to encourage advocacy for different facets through
collaboration.

For example:

* the [**core** team](teams/core.yml) has authority over the [RFC
  process][rfcs-repo] and associated design principles, but cannot directly
  push to the the [Concourse repo][concourse-repo].
* the [**maintainers** team](teams/maintainers.yml) has authority over the
  [Concourse repo][concourse-repo] and submits RFCs to develop a roadmap that
  aligns with Concourse's core design principles.
* the **core** team engages with the **maintainers** team to ensure new
  proposals do not introduce unnecessary risk or become a maintenance burden.
* the **maintainers** team then guides the planning and implementation of the
  proposal through pull requests to the Concourse repo.

Teams may split off from larger teams as more of these boundaries are
discovered. Careful attention should be paid to teams with too many
responsibilities - there may be significant facets being neglected.

[rfcs-repo]: https://github.com/concourse/rfcs
[concourse-repo]: https://github.com/concourse/concourse


### Individual Contributors

Individual contributors are listed under `./contributors`. Pull requests will
be reviewed by members of the **community** team. Feel free to submit one at
any time!

The name of the file should match your github handle. Each
`./contributors/*.yml` file has the following fields:

* `name` - the contributor's real name, or an alias if they would rather not
  share.
* `github` - the contributor's GitHub login
* `discord` - the contributor's Discord username + number, e.g. `foo#123`
* `repos` - map from repo name to permission to grant for the user. this should
  only be used for bot accounts; in general repo permissions should be done
  through teams.

Each contributor will be granted membership of the Concourse GitHub
organization. This does not grant much on its own; repository access for
example is determined through teams.

> Note: the Discord attribute is not used at the moment, but it may be helpful
> in the future to have someplace that correlates these different identities.



### Teams

Teams are listed under `./teams`. Pull requests will be reviewed by the
**community** team, who will further request reviews from all affected teams or
individuals. (This can probably be automated at some point.)

Each `./teams/*.yml` file has the following fields:

* `name` - a name for the team, stylized in lowercase.
* `purpose` - a brief description of the team's focus.
* `responsibilities` - a list of the team's discrete responsibilities, or a
  link to where they can be found.
* `members` - a list of contributors to add to the team, e.g. `foo` for
  `./contributors/foo.yml`.
* `repos` - a list of GitHub repositories for the team to be added to.

Each team must have a stated purpose summarizing its goals.

Each team is also responsible for maintaining a list of its responsibilities.
(No need to list that one.) Doing so clarifies the scope of a team for
newcomers and makes it easier to tell when a team is overloaded and could
benefit from being divided or reorganized.

Each team lists its members which correspond to filenames under
`./contributors` (without the `.yml`).

Each team lists GitHub repositories for which the team will be granted
the [Maintain permission][permissions].

Each team is responsible for determining the best way for the team to operate,
though it is strongly encouraged that each team work in the open, either on
GitHub or somewhere easy to access, to the extent that doing so is beneficial
to the team and to the community. (For example, teams may choose to use a
private discussion area to handle sensitive matters.)

Suggestion: team processes can be defined in a new repository managed
exclusively by the team. The team repository can be created via submitting a PR
to this repo. See [Repositories](#repositories).

#### Voting

Decisions are reached through consensus among the team members through a 66%+
supermajority unless stated otherwise through the team's own processes.
(Implementation of said process would require a 66% supermajority.)

Voting can be expressed through pull request review, leaving a comment, or
through some other form of record - ideally permanent.

Teams are not required to have designated leaders. Teams may choose to
designate a leader and define their role and responsibilities through a vote
amongst the team.

[permissions]: https://docs.github.com/en/github/setting-up-and-managing-organizations-and-teams/repository-permission-levels-for-an-organization

#### Joining a Team

To propose the addition of a team member (either yourself or on behalf of
someone else), submit a PR that adds them as a contributor (if needed) and
lists them as a member of the desired team.

There are no specific qualifications for joining a team - being accepted into a
team is entirely subjective and the barrier to entry will vary from team to
team. Applications with no prior context or trust to build upon will almost
certainly be rejected.

Pull requests that add someone to a team require enough approving reviewers to
pass the [voting process](#voting). The **community** team is responsible for
determining the required votes and merging when they have all been acquired.


#### Creating a new Team

New teams may be formed at any time by submitting a PR. A team with only one
member is probably not a good sign, so try to recruit folks during this stage.

If a new team is being created to split off from a larger team, you will have
to negotiate ownership of the relevant repos and ideally move them entirely to
the new team. This will obviously require approval from the original team.


### Repositories

Repositories are listed under `./repos`. Pull requests will be reviewed by the
**infrastructure** team.

Each `./repos/*.yml` file has the following fields:

* `name` - a name for the repository.
* `description` - a description for the repository.
* `topics` - topics to set for the repository.
* `homepage_url` - a website (if any) associated to the repository.
* `has_issues` - whether the repository has Issues enabled (default `false`).
* `has_projects` - whether the repository has Projects enabled (default
  `false`).
* `has_wiki` - whether the repository has the Wiki enabled (default `false`).
* `pages` - GitHub pages configuration:
  * `branch` - the branch to build.
  * `path` - the path to serve (default `/`).
  * `cname` - an optional CNAME to set for the website.
* `branch_protection` - a list of branch protection settings:
  * `pattern` - branch name pattern to match.
  * `allows_deletions` - whether the branches can be deleted.
  * `required_checks` - required status checks for PRs to be merged.
  * `strict_checks` - require branches to be up-to-date before merging.
  * `required_reviews` - number of approved reviews required for PRs to be
    merged.
  * `dismiss_stale_reviews` - dismiss reviews when new commits are pushed.
  * `require_code_owner_reviews` - require approval from code owners for PRs
    which affect files with designated owners.
* `deploy_keys` - a list of [deploy keys] to add to the repo
  * `title` - a title for the key
  * `public_key` - the public key
  * `writable` - whether the key can push to the repo

All repositories have [vulnerability alerts] enabled.

All repositories are configured to [delete branches] once their PR is merged.

All repositories will be archived upon deletion from this repo (instead of
being deleted). Permanent deletion must be done manually by a member of the
**infrastructure** team.

[vulnerability alerts]: https://docs.github.com/en/github/managing-security-vulnerabilities/about-alerts-for-vulnerable-dependencies
[delete branches]: https://docs.github.com/en/github/administering-a-repository/managing-the-automatic-deletion-of-branches
[deploy keys]: https://docs.github.com/en/developers/overview/managing-deploy-keys


## Amending the Governance Model

> Frankly, I am more used to solving computer problems than human problems, so
> this process may be naive, it may feel too rigid, or it may feel completely
> ambiguous. Nothing here is set in stone. Please improve it as necessary and
> remove this disclaimer once we feel more confident. - **@vito**

Pull requests to this process (`README.md`) will be reviewed by the
**core** team.


## Enforcing the Governance Model

The configuration in this repository is applied automatically via Terraform.

In addition to the Terraform configuration, the state of the entire GitHub
organization can be tested against the desired state via `go test`. This test
suite will also detect any 'extra' configuration like untracked repositories,
unknown teams, and extra repository collaborators.


### Applying Changes

To apply these changes you must be an Owner of the Concourse GitHub
organization.

Set the `github_token` var and run `terraform apply`:

```sh
$ terraform init # once
$ echo '{"github_token":"..."}' > .auto.tfvars.json
$ terraform apply
```

This token must have *admin:org* and *repo* scopes.


### Testing Actual vs. Desired State

Tests are included which will verify that all permissions in the relevant
services reflect the configuration in the repository.

Running the tests requires a `$GITHUB_TOKEN` to be set.

```sh
$ export GITHUB_TOKEN="$(jq -r .github_token .auto.tfvars.json)"
$ go test
```

Test failures must be addressed immediately as they may indicate abuse, though
laziness or ignorance of this process is more likely.


### GitHub Organization Settings

This governance model requires that organization members have extremely limited
[privileges][member-privileges]. Unfortunately these can't currently be set by
Terraform, so I'm documenting them here for good measure.

The following settings are required for any of this to make sense:

* **Base permissions** must be "None" so that organization membership does not
  grant visibility of private repositories (if any exist) and repository
  access is determined exclusively through teams.
* **Repository creation** and **Pages creation** must be disabled for both
  Public and Private so that all repository management shall be done through
  this repo.
* **Allow members to create teams** must be disabled so that all team
  administration shall be done through this repo.

Additionally, repository admin permissions should be restricted. No team will
ever be an 'admin' at the repo level, so this should never come up, but we can
prevent further damage if someone does manage to escalate:

* **Allow members to change repository visibilities for this organization**
  should be disabled.
* **Allow members to delete or transfer repositories for this organization**
  should be disabled.
* **Allow members to delete issues for this organization** should be disabled.

These settings probably won't have much impact:

* **Allow forking of private repositories** should be disabled just to keep
  access tidy.
* **Allow users with read access to create discussions** is confusingly under
  the 'Admin repository permissions' heading but sounds rather innocuous, so it
  can be left checked.

*More settings may appear on the member privileges page at some point. Please
update the above listing if/when this does occur.*

[member-privileges]: https://github.com/organizations/concourse/settings/member_privileges
