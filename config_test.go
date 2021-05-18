package governance_test

import (
	"os"
	"testing"

	"github.com/concourse/governance"
	"github.com/stretchr/testify/require"
)

func TestMemberEmails(t *testing.T) {
	config, err := governance.LoadConfig(os.DirFS("."))
	require.NoError(t, err)

	for _, team := range config.Teams {
		for _, member := range team.Members(config) {
			if team.RequiresEmail {
				require.NotEmpty(t,
					member.Email,
					"team %s requires email, but member %s does not have one",
					team.Name, member.Name,
				)
			}
		}
	}
}
