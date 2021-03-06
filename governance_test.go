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
		for _, repo := range actual.Repos {
			t.Run(repo.Name, func(t *testing.T) {
				t.Run("has no collaborators", func(t *testing.T) {
					require.Empty(t, repo.DirectCollaborators)
				})
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
