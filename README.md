# Concourse Governance

This repository codifies the roles and permissions within the Concourse project
and contains the code to apply them via [Terraform](https://www.terraform.io/).


## Individual Contributors

Individual contributors are listed under `./contributors` - feel free to submit
a PR adding yourself!

Each contributor will be granted mermbership of the Concourse GitHub
organization. This does not grant much on its own; repository access for
example is determined through teams.

Each `./contributors/*.yml` file has the following fields:

* `name` - the contributor's real name, or an alias if they would rather not
  share.
* `github` - the contributor's GitHub login
* `discord` - the contributor's Discord username + number, e.g. `foo#123`
* `admin` - whether the contributor will be an admin of the organization.

Pull requests to `./contributors` will be reviewed by members of the
**@concourse/community** team.


## Teams

Teams are listed under `./teams`. Each team must have a stated purpose
summarizing its goals.

New teams may be formed at any time by submitting a PR. A team with only one
member is probably not a good sign, so try to recruit folks during this stage.

Each team lists its members which correspond to filenames under
`./contributors` (without the `.yml`).

Each team lists GitHub repositories for which the team will be granted
the [`maintain` permission][permissions].

Teams do not have designated leadership, though there may be a reason to add
this someday. It is assumed that decisions are made through consensus among the
team.

Each team is responsible for determining the best way for the team to operate,
though it a requirement that each team work in the open, either on GitHub or
somewhere easy to access.

Each team is responsible for maintaining a list of its responsibilities. (No
need to list that one.) This clarifies the scope of a team for newcomers and to
makes it easier to notice when a team is overloaded and could benefit from
being divided or reorganized.

Each `./teams/*.yml` file has the following fields:

* `name` - a name for the team, stylized in lowercase.
* `purpose` - a brief description of the team's focus.
* `responsibilities` - a list of the team's discrete responsibilities, or a
  link to where they can be found.
* `members` - a list of contributors to add to the team, e.g. `foo` for
  `./contributors/foo.yml`.
* `repos` - a list of GitHub repositories for the team to be added to.

Pull requests to `./teams` will be reviewed by the **@concourse/community**
team, the affected teams, and any members being added to a team.

[permissions]: https://docs.github.com/en/github/setting-up-and-managing-organizations-and-teams/repository-permission-levels-for-an-organization


## Repositories

Repositories are listed under `./repos`.

Pull requests to `./repos` will be reviewed by the
**@concourse/infrastructure** team.

(TODO)


## Amending the Governance Model

> Frankly, I am more used to solving computer problems than human problems, so
> this may be naive, it may feel too rigid, or it may feel completely
> ambiguous. Nothing here is set in stone. Please improve it as necessary and
> remove this disclaimer once we feel more confident. - **@vito**

Pull requests to this process (`README.md`) will be reviewed by the
**@concourse/community** team.


## Applying Changes

To apply these changes you must be an Owner of the Concourse GitHub
organization.

Set the `github_token` var and run `terraform apply`:

```sh
$ terraform init # once
$ echo '{"github_token":"..."}' > .auto.tfvars.json
$ terraform apply
```

This token must have *admin:org* and *repo* scopes.


## Testing Actual vs. Desired State

Tests are included which will verify that all permissions in the relevant
services reflect the configuration in the repository.

Running the tests requires a `$GITHUB_TOKEN` to be set.

```sh
$ export GITHUB_TOKEN="$(jq -r .github_token .auto.tfvars.json)"
$ go test
```

Test failures must be addressed immediately as they may indicate abuse, though
laziness or ignorance of this process is more likely.


## GitHub Organization Settings

This governance model requires that organization members have extremely limited
[privileges][member-privileges]. Unfortunately these can't currently be set by
Terraform, so I'm documenting them here for good measure.

The following settings are required for any of this to make sense:

* **Base permissions** must be "None" so that organization membership does not
  grant visibility of private repositories (if any exist) and repository
  access is determined exclusively through teams.
* **Repository creation** must be disabled for both Public and Private so that
  all repository management shall be done through this repo.
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

* **Pages creation** can be allowed for Public repositories.
* **Allow forking of private repositories** should be disabled just to keep
  access tidy.
* **Allow users with read access to create discussions** is confusingly under
  the 'Admin repository permissions' heading but sounds rather innocuous, so it
  can be left checked.

*More settings may appear on the member privileges page at some point. Please
update the above listing if/when this does occur.*

[member-privileges]: https://github.com/organizations/concourse/settings/member_privileges
