package main

import (
	"log"
	"os"
	"sort"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/concourse/governance"
	"go.uber.org/zap"
)

// Concourse Discord server ID
const guildID = "219899946617274369"

// suffix applied to team-based roles
const teamRoleSuffix = "-team"

// defaults copied from newly created role; may be worth tuning later
var teamRoleBasePermissions = []string{
	"VIEW_CHANNEL",
	"CREATE_INSTANT_INVITE",
	"CHANGE_NICKNAME",
	"SEND_MESSAGES",
	"EMBED_LINKS",
	"ATTACH_FILES",
	"ADD_REACTIONS",
	"USE_EXTERNAL_EMOJIS",
	"MENTION_EVERYONE",
	"READ_MESSAGE_HISTORY",
	"SEND_TTS_MESSAGES",
	"CONNECT",
	"SPEAK",
	"STREAM",
	"USE_VAD",
}

func main() {
	logger, err := zap.NewDevelopment(zap.IncreaseLevel(zap.InfoLevel))
	if err != nil {
		log.Fatalln("zap:", err)
	}

	defer logger.Sync()

	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		logger.Fatal("no $DISCORD_TOKEN provided")
	}

	session, err := discordgo.New("Bot " + token)
	if err != nil {
		logger.Fatal("failed to initialize discord", zap.Error(err))
	}

	config, err := governance.LoadConfig(os.DirFS("."))
	if err != nil {
		logger.Fatal("failed to load config", zap.Error(err))
	}

	userIDToName := map[string]string{}

	discordMembers := []*discordgo.Member{}
	after := ""
	limit := 1000
	for {
		members, err := session.GuildMembers(guildID, after, limit)
		if err != nil {
			logger.Fatal("failed to get members", zap.Error(err))
		}

		for _, member := range members {
			discordMembers = append(discordMembers, member)
			userIDToName[member.User.ID] = member.User.String()
			after = member.User.ID
		}

		if len(members) < limit {
			break
		}
	}

	actualUserRoles := map[string]map[string]bool{}
	for _, member := range discordMembers {
		actualRoles, found := actualUserRoles[member.User.ID]
		if !found {
			actualRoles = map[string]bool{}
			actualUserRoles[member.User.ID] = actualRoles
		}

		for _, role := range member.Roles {
			actualRoles[role] = true
		}
	}

	discordRoles, err := session.GuildRoles(guildID)
	if err != nil {
		logger.Fatal("failed to get roles", zap.Error(err))
	}

	roleIDToName := map[string]string{}
	for _, role := range discordRoles {
		roleIDToName[role.ID] = role.Name
	}

	var roleOrder []*discordgo.Role
	var teams []governance.Team
	for _, team := range config.Teams {
		if team.Legacy {
			// don't set up roles for legacy teams
			continue
		}

		teams = append(teams, team)
	}

	sort.Sort(byPriority(teams))

	desiredUserRoles := map[string]map[string]bool{}
	for position, team := range teams {
		roleName := team.Name + teamRoleSuffix

		logger := logger.With(
			zap.String("team", team.Name),
			zap.String("role", roleName),
		)

		var teamRole *discordgo.Role
		for _, role := range discordRoles {
			if role.Name == roleName {
				teamRole = role
				break
			}
		}

		if teamRole != nil {
			logger.Debug("role already exists")
		} else {
			logger.Info("creating role")

			teamRole, err = session.GuildRoleCreate(guildID)
			if err != nil {
				logger.Fatal("failed to create role", zap.Error(err))
			}

			roleIDToName[teamRole.ID] = roleName
		}

		var permissions int64
		for _, permission := range append(teamRoleBasePermissions, team.Discord.Permissions...) {
			bits, found := governance.DiscordPermissions[permission]
			if !found {
				logger.Error("unknown permission", zap.String("permission", permission))
			}

			permissions |= bits
		}

		teamRole, err = session.GuildRoleEdit(
			guildID,
			teamRole.ID,
			roleName,
			team.Discord.Color,
			true, // hoist
			permissions,
			true, // mentionable
		)
		if err != nil {
			log.Fatal("failed to update role", zap.Error(err))
		}

		teamRole.Position = position + 1
		roleOrder = append(roleOrder, teamRole)

		for _, member := range team.Members {
			logger := logger.With(
				zap.String("member", member),
			)

			contributor, found := config.Contributors[member]
			if !found {
				logger.Debug("contributor not found")
				continue
			}

			if contributor.Discord == "" {
				logger.Debug("contributor has no discord user")
				continue
			}

			logger = logger.With(
				zap.String("user", contributor.Discord),
			)

			var discordMember *discordgo.Member
			for _, member := range discordMembers {
				if member.User.String() == contributor.Discord {
					discordMember = member
				}
			}

			if discordMember == nil {
				logger.Debug("user is not a member of the server")
				continue
			}

			desiredRoles, found := desiredUserRoles[discordMember.User.ID]
			if !found {
				desiredRoles = map[string]bool{}
				desiredUserRoles[discordMember.User.ID] = desiredRoles
			}

			desiredRoles[teamRole.ID] = true
		}
	}

	_, err = session.GuildRoleReorder(guildID, roleOrder)
	if err != nil {
		logger.Error("failed to reorder roles", zap.Error(err))
	}

	for userID, desiredRoles := range desiredUserRoles {
		actualRoles, found := actualUserRoles[userID]
		if !found {
			actualRoles = map[string]bool{}
		}

		for roleID := range desiredRoles {
			logger := logger.With(
				zap.String("user", userIDToName[userID]),
				zap.String("role", roleIDToName[roleID]),
			)

			if actualRoles[roleID] {
				logger.Debug("user already has role")
			} else {
				logger.Info("adding role to user")

				err = session.GuildMemberRoleAdd(guildID, userID, roleID)
				if err != nil {
					logger.Fatal("failed to add role", zap.Error(err))
				}
			}
		}
	}

	for userID, actualRoles := range actualUserRoles {
		desiredRoles, found := desiredUserRoles[userID]
		if !found {
			desiredRoles = map[string]bool{}
		}

		for roleID := range actualRoles {
			roleName := roleIDToName[roleID]

			logger := logger.With(
				zap.String("user", userIDToName[userID]),
				zap.String("role", roleName),
			)

			if !strings.HasSuffix(roleName, teamRoleSuffix) {
				logger.Debug("ignoring non-team role")
				continue
			}

			if !desiredRoles[roleID] {
				logger.Info("removing team role from user")

				err = session.GuildMemberRoleRemove(guildID, userID, roleID)
				if err != nil {
					logger.Fatal("failed to remove role", zap.Error(err))
				}
			}
		}
	}
}

type byPriority []governance.Team

func (teams byPriority) Len() int { return len(teams) }

func (teams byPriority) Less(i, j int) bool {
	return teams[i].Discord.Priority < teams[j].Discord.Priority
}

func (teams byPriority) Swap(i, j int) {
	teams[i], teams[j] = teams[j], teams[i]
}
