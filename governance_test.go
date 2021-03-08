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
		for _, repo := range desired.Repos {
			actualRepo, found := actual.Repo(repo.Name)
			require.True(t, found, "repo does not exist")

			t.Run(actualRepo.Name, func(t *testing.T) {
				t.Run("has matching configuration", func(t *testing.T) {
					require.Equal(t, repo.Description, actualRepo.Description)
					require.Equal(t, repo.Topics, actualRepo.Topics)
					require.Equal(t, repo.HasIssues, actualRepo.HasIssues)
					require.Equal(t, repo.HasProjects, actualRepo.HasProjects)
					require.Equal(t, repo.HasWiki, actualRepo.HasWiki)
				})

				t.Run("has no collaborators", func(t *testing.T) {
					require.Empty(t, actualRepo.DirectCollaborators)
				})
			})
		}

		for _, repo := range actual.Repos {
			_, found := desired.Repo(repo.Name)
			if found {
				continue
			}

			t.Run(repo.Name, func(t *testing.T) {
				t.Error("repo should not exist")
			})
		}
	})

	t.Run("teams", func(t *testing.T) {
		for _, team := range desired.Teams {
			t.Run(team.Name, func(t *testing.T) {
				actualTeam, found := actual.Team(team.Name)
				require.True(t, found, "team does not exist")

				t.Run("members", func(t *testing.T) {
					require.ElementsMatch(t, team.Members, actualTeam.Members)
				})

				t.Run("repos", func(t *testing.T) {
					require.ElementsMatch(t, team.Repos, actualTeam.Repos)
				})
			})
		}

		for _, team := range actual.Teams {
			_, found := desired.Team(team.Name)
			if found {
				continue
			}

			t.Run(team.Name, func(t *testing.T) {
				t.Error("team should not exist")
			})
		}
	})
}
