package delta

import (
	"fmt"
	"sort"

	"github.com/concourse/governance"
)

func Diff(config *governance.Config, discord Discord) ([]Delta, error) {
	var deltas []Delta

	members, err := discord.Members()
	if err != nil {
		return nil, fmt.Errorf("get members: %w", err)
	}

	userIDToName := map[string]string{}
	nameToUserID := map[string]string{}

	actualUserRoles := map[string]map[string]bool{}
	desiredUserRoles := map[string]map[string]bool{}

	for _, member := range members {
		userIDToName[member.ID] = member.Name
		nameToUserID[member.Name] = member.ID

		actualRoles, found := actualUserRoles[member.ID]
		if !found {
			actualRoles = map[string]bool{}
			actualUserRoles[member.ID] = actualRoles
		}

		for _, role := range member.RoleNames {
			actualRoles[role] = true
		}
	}

	actualRoles, err := discord.Roles()
	if err != nil {
		return nil, fmt.Errorf("get members: %w", err)
	}

	sort.Sort(sort.Reverse(byPosition(actualRoles)))

	roleIDToName := map[string]string{}
	roleNameToID := map[string]string{}
	for _, role := range actualRoles {
		roleIDToName[role.ID] = role.Name
		roleNameToID[role.Name] = role.ID
	}

	var teams []governance.Team
	for _, team := range config.Teams {
		teams = append(teams, team)
	}

	sort.Sort(byPriority(teams))

	roleOrder := make([]string, len(teams))
	teamRoles := map[string]bool{}
	stickyRoles := map[string]bool{}
	for position, team := range teams {
		roleName := team.Discord.Role
		if roleName == "" {
			roleName = team.Name + "-team"
		}

		roleOrder[position] = roleName
		teamRoles[roleName] = true

		if team.Discord.Sticky {
			stickyRoles[roleName] = true
		}

		var roleExists bool
		var existingRole DiscordRole
		for _, role := range actualRoles {
			if role.Name == roleName {
				roleExists = true
				existingRole = role
				break
			}
		}

		permissionSet := append(
			team.Discord.AddedPermissions,
			governance.TeamRoleBasePermissions...,
		)

		permissions, err := permissionSet.Permissions()
		if err != nil {
			return nil, err
		}

		if !roleExists {
			deltas = append(deltas, DeltaRoleCreate{
				RoleName:    roleName,
				Color:       team.Discord.Color,
				Permissions: permissions,
			})
		} else if existingRole.Color != team.Discord.Color || existingRole.Permissions != permissions {
			deltas = append(deltas, DeltaRoleEdit{
				RoleID:      existingRole.ID,
				RoleName:    roleName,
				Color:       team.Discord.Color,
				Permissions: permissions,
			})
		}

		for _, contributor := range team.Members(config) {
			if contributor.Discord == "" {
				continue
			}

			userID, found := nameToUserID[contributor.Discord]
			if !found {
				continue
			}

			desiredRoles, found := desiredUserRoles[userID]
			if !found {
				desiredRoles = map[string]bool{}
				desiredUserRoles[userID] = desiredRoles
			}

			desiredRoles[roleName] = true
		}
	}

	actualRoleOrder := []string{}
	for _, role := range actualRoles {
		if teamRoles[role.Name] {
			actualRoleOrder = append(actualRoleOrder, role.Name)
		}
	}

	sameOrder := true
	if len(roleOrder) != len(actualRoleOrder) {
		sameOrder = false
	} else {
		for i := range roleOrder {
			if roleOrder[i] != actualRoleOrder[i] {
				sameOrder = false
			}
		}
	}

	if !sameOrder {
		deltas = append(deltas, DeltaRolePositions(roleOrder))
	}

	var addUserRoles []DeltaUserAddRole
	for userID, desiredRoles := range desiredUserRoles {
		actualRoles, found := actualUserRoles[userID]
		if !found {
			actualRoles = map[string]bool{}
		}

		for roleName := range desiredRoles {
			if actualRoles[roleName] {
				continue
			}

			addUserRoles = append(addUserRoles, DeltaUserAddRole{
				UserID:   userID,
				UserName: userIDToName[userID],
				RoleName: roleName,
			})
		}
	}

	sort.Sort(byAddRole(addUserRoles))
	for _, v := range addUserRoles {
		deltas = append(deltas, v)
	}

	var removeUserRoles []DeltaUserRemoveRole
	for userID, actualRoles := range actualUserRoles {
		desiredRoles, found := desiredUserRoles[userID]
		if !found {
			desiredRoles = map[string]bool{}
		}

		for roleName := range actualRoles {
			if stickyRoles[roleName] {
				// removing the 'contributors' role is more disruptive than it's worth,
				// since almost all of them predate this process and don't have their
				// Discord associated to a GitHub account.
				continue
			}

			if !teamRoles[roleName] {
				// only team roles are removed.
				//
				// removing other roles can be done once there's a general process for
				// managing Discord roles - for now it's manual through the community
				// team.
				continue
			}

			if !desiredRoles[roleName] {
				removeUserRoles = append(removeUserRoles, DeltaUserRemoveRole{
					UserID:   userID,
					UserName: userIDToName[userID],
					RoleName: roleName,
				})
			}
		}
	}

	sort.Sort(byRemoveRole(removeUserRoles))
	for _, v := range removeUserRoles {
		deltas = append(deltas, v)
	}

	return deltas, nil
}

type byPosition []DiscordRole

func (roles byPosition) Len() int { return len(roles) }

func (roles byPosition) Swap(i, j int) {
	roles[i], roles[j] = roles[j], roles[i]
}

func (roles byPosition) Less(i, j int) bool {
	return roles[i].Position < roles[j].Position
}

type byAddRole []DeltaUserAddRole

func (deltas byAddRole) Len() int { return len(deltas) }

func (deltas byAddRole) Less(i, j int) bool {
	return deltas[i].UserID < deltas[j].UserID ||
		deltas[i].RoleName < deltas[j].RoleName
}

func (deltas byAddRole) Swap(i, j int) {
	deltas[i], deltas[j] = deltas[j], deltas[i]
}

type byRemoveRole []DeltaUserRemoveRole

func (deltas byRemoveRole) Len() int { return len(deltas) }

func (deltas byRemoveRole) Less(i, j int) bool {
	return deltas[i].UserID < deltas[j].UserID ||
		deltas[i].RoleName < deltas[j].RoleName
}

func (deltas byRemoveRole) Swap(i, j int) {
	deltas[i], deltas[j] = deltas[j], deltas[i]
}

type byPriority []governance.Team

func (teams byPriority) Len() int { return len(teams) }

func (teams byPriority) Less(i, j int) bool {
	return teams[i].Discord.Priority < teams[j].Discord.Priority
}

func (teams byPriority) Swap(i, j int) {
	teams[i], teams[j] = teams[j], teams[i]
}
