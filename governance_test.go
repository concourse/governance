package governance_test

import (
	"os"
	"sort"
	"testing"

	"github.com/concourse/governance"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGitHub(t *testing.T) {
	config, err := governance.LoadConfig(os.DirFS("."))
	require.NoError(t, err)

	desired := config.DesiredGitHubState()

	actual, err := governance.LoadGitHubState("concourse")
	require.NoError(t, err)

	t.Run("members", func(t *testing.T) {
		for _, member := range desired.Members {
			actualMember, found := actual.Member(member.Login)
			if assert.True(t, found, "%s should be a member of the organization, but is not", member.Login) {
				assert.Equal(t, member.Role, actualMember.Role, "%s has wrong role", member.Login)
			}
		}

		for _, member := range actual.Members {
			_, found := desired.Member(member.Login)
			assert.True(t, found, "%s should not be a member", member.Login)
		}
	})

	t.Run("repos", func(t *testing.T) {
		for _, d := range desired.Repos {
			desiredRepo := d

			t.Run(desiredRepo.Name, func(t *testing.T) {
				actualRepo, found := actual.Repo(desiredRepo.Name)
				require.True(t, found, "repo does not exist")

				t.Run("has matching configuration", func(t *testing.T) {
					require.Equal(t, desiredRepo.Description, actualRepo.Description, "description")
					require.Equal(t, desiredRepo.IsPrivate, actualRepo.IsPrivate, "privacy")
					require.ElementsMatch(t, desiredRepo.Topics, actualRepo.Topics, "topics")
					require.Equal(t, desiredRepo.HomepageURL, actualRepo.HomepageURL, "homepage URL")
					require.Equal(t, desiredRepo.HasIssues, actualRepo.HasIssues, "has issues")
					require.Equal(t, desiredRepo.HasProjects, actualRepo.HasProjects, "has projects")
					require.Equal(t, desiredRepo.HasWiki, actualRepo.HasWiki, "has wiki")
					require.ElementsMatch(t, desiredRepo.DirectCollaborators, actualRepo.DirectCollaborators, "collaborators")
				})

				t.Run("has correct branch protection", func(t *testing.T) {
					for _, rule := range desiredRepo.BranchProtectionRules {
						sort.Strings(rule.RequiredStatusCheckContexts)
					}

					for _, rule := range actualRepo.BranchProtectionRules {
						sort.Strings(rule.RequiredStatusCheckContexts)
					}

					require.ElementsMatch(t, desiredRepo.BranchProtectionRules, actualRepo.BranchProtectionRules, "branch protection")
				})

				t.Run("belongs to a team", func(t *testing.T) {
					var belongs bool
					for _, team := range desired.Teams {
						for _, repo := range team.Repos {
							if repo.Name == desiredRepo.Name {
								belongs = true
							}
						}
					}

					require.True(t, belongs, "does not belong to any team")
				})
			})
		}

		for _, a := range actual.Repos {
			actualRepo := a

			_, found := desired.Repo(actualRepo.Name)
			if found {
				continue
			}

			t.Run(actualRepo.Name, func(t *testing.T) {
				t.Error("repo is not in configuration")
			})
		}
	})

	t.Run("teams", func(t *testing.T) {
		for _, d := range desired.Teams {
			desiredTeam := d

			t.Run(desiredTeam.Name, func(t *testing.T) {
				actualTeam, found := actual.Team(desiredTeam.Name)
				require.True(t, found, "team does not exist")

				t.Run("members", func(t *testing.T) {
					for _, member := range desiredTeam.Members {
						actualMember, found := actualTeam.Member(member.Login)
						if assert.True(t, found, "%s should be a member of the %s team, but is not", member.Login, desiredTeam.Name) {
							assert.Equal(t, member.Role, actualMember.Role, "%s has wrong role", member.Login)
						}
					}

					for _, member := range actualTeam.Members {
						_, found := desiredTeam.Member(member.Login)
						assert.True(t, found, "%s should not be a member of the %s team", member.Login, desiredTeam.Name)
					}
				})

				t.Run("repos", func(t *testing.T) {
					require.ElementsMatch(t, desiredTeam.Repos, actualTeam.Repos)
				})
			})
		}

		for _, a := range actual.Teams {
			actualTeam := a

			_, found := desired.Team(actualTeam.Name)
			if found {
				continue
			}

			t.Run(actualTeam.Name, func(t *testing.T) {
				t.Error("team should not exist")
			})
		}
	})
}
