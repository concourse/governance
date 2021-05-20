package governance_test

import (
	"os"
	"testing"

	"github.com/concourse/governance"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const domain = "concourse-ci.org"

func TestMailgun(t *testing.T) {
	config, err := governance.LoadConfig(os.DirFS("."))
	require.NoError(t, err)

	desired := config.DesiredMailgunState(domain)

	actual, err := governance.LoadMailgunState(domain)
	require.NoError(t, err)

	for _, desiredRoute := range desired.Routes {
		t.Run(desiredRoute.Description, func(t *testing.T) {
			var found bool
			for _, actualRoute := range actual.Routes {
				if actualRoute.Description == desiredRoute.Description {
					found = true
					assert.Equal(t, desiredRoute.Expression, actualRoute.Expression)
					assert.ElementsMatch(t, desiredRoute.Actions, actualRoute.Actions)
					break
				}
			}

			assert.True(t, found, "route is not configured")
		})
	}

	for _, actualRoute := range actual.Routes {
		t.Run(actualRoute.Description, func(t *testing.T) {
			var found bool
			for _, desiredRoute := range desired.Routes {
				if desiredRoute.Description == actualRoute.Description {
					found = true
					break
				}
			}

			assert.True(t, found, "route is not desired")
		})
	}
}
