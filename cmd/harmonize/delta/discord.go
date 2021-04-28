package delta

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

type Discord interface {
	Members() ([]DiscordMember, error)
	Roles() ([]DiscordRole, error)
}

type DiscordMember struct {
	ID        string
	Name      string
	RoleNames []string
}

type DiscordRole struct {
	ID          string
	Name        string
	Color       int
	Permissions int64
	Position    int
}

type discord struct {
	session *discordgo.Session
	guildID string
}

func NewDiscord(guildID, token string) (Discord, error) {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, fmt.Errorf("init discordgo: %w", err)
	}

	return &discord{
		session: session,
		guildID: guildID,
	}, nil
}

func (discord *discord) Members() ([]DiscordMember, error) {
	discordMembers := []DiscordMember{}

	discordRoles, err := discord.session.GuildRoles(discord.guildID)
	if err != nil {
		return nil, fmt.Errorf("get guild roles: %w", err)
	}

	after := ""
	limit := 1000
	for {
		members, err := discord.session.GuildMembers(discord.guildID, after, limit)
		if err != nil {
			return nil, fmt.Errorf("get guild members: %w", err)
		}

		for _, member := range members {
			roleNames := make([]string, len(member.Roles))

		dance:
			for i, roleID := range member.Roles {
				for _, role := range discordRoles {
					if role.ID == roleID {
						roleNames[i] = role.Name
						break dance
					}
				}
			}

			discordMembers = append(discordMembers, DiscordMember{
				ID:        member.User.ID,
				Name:      member.User.String(),
				RoleNames: roleNames,
			})

			after = member.User.ID
		}

		if len(members) < limit {
			break
		}
	}

	return discordMembers, nil
}

func (discord *discord) Roles() ([]DiscordRole, error) {
	discordRoles, err := discord.session.GuildRoles(discord.guildID)
	if err != nil {
		return nil, fmt.Errorf("get guild roles: %w", err)
	}

	roles := make([]DiscordRole, len(discordRoles))
	for i, r := range discordRoles {
		roles[i] = DiscordRole{
			ID:          r.ID,
			Name:        r.Name,
			Color:       r.Color,
			Permissions: r.Permissions,
			Position:    r.Position,
		}
	}

	return roles, nil
}
