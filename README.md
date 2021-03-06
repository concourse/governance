# Concourse Governance

This repository codifies the roles and permissions within the Concourse
project. It currently maps to GitHub teams and repositories but could
theoretically be used to assign equivalent roles in other places too (e.g.
Discord, mailing lists).


## `./people`

The `./people` directory contains one file per individual contributor. Each
contributor will be granted membership to the Concourse GitHub organization.

Each person `.yml` file has the following schema:

* `name` - the user's real name, or an alias if they would rather not share.
* `github` - the user's GitHub login
* `discord` - the user's Discord username + number, e.g. `foo#123`
* `admin` - if set to `true`, the user will be an owner of the GitHub
  organization and an admin in Discord.

Note that organization membership does not grant write access to any
repositories. Repository access control is only determined by teams.

The obvious exception is `admin: true`, which should be limited to to members
of the `infrastructure` team.


## `./teams`

The `./teams` directory contains one file per team. A team is just a
subdivision of contributors with a stated purpose, ideally narrow enough in
scope that teams may be formed organically or disbanded if they become no
longer necessary.

Each team `.yml` file has the following schema:

* `name` - a name for the team, stylized in lowercase.
* `purpose` - a brief description of the team's focus.
* `members` - a list of people to add to the team, identified by their filename
  (without the `.yml`).
* `repos` - a list of GitHub repositories for the team to be added to.

Teams will be added to repositories with the `maintain` role.
