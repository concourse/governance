package governance_test

import (
	"os"
	"testing"

	"github.com/concourse/governance"
	"github.com/stretchr/testify/require"
)

func TestGitHub(t *testing.T) {
	config, err := governance.LoadConfig(os.DirFS("."))
	require.NoError(t, err)

	desired := config.DesiredGitHubState()

	actual, err := governance.LoadGitHubState("concourse")
	require.NoError(t, err)

	t.Run("members", func(t *testing.T) {
		require.ElementsMatch(t, desired.Members, actual.Members)
	})

	t.Run("repos", func(t *testing.T) {
		for _, d := range desired.Repos {
			desiredRepo := d

			t.Run(desiredRepo.Name, func(t *testing.T) {
				actualRepo, found := actual.Repo(desiredRepo.Name)
				require.True(t, found, "repo does not exist")

				t.Run("has matching configuration", func(t *testing.T) {
					require.Equal(t, desiredRepo.Description, actualRepo.Description)
					require.Equal(t, desiredRepo.Topics, actualRepo.Topics)
					require.Equal(t, desiredRepo.HasIssues, actualRepo.HasIssues)
					require.Equal(t, desiredRepo.HasProjects, actualRepo.HasProjects)
					require.Equal(t, desiredRepo.HasWiki, actualRepo.HasWiki)
				})

				t.Run("has no collaborators", func(t *testing.T) {
					require.Empty(t, actualRepo.DirectCollaborators)
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
					require.ElementsMatch(t, desiredTeam.Members, actualTeam.Members)
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
