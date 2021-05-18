package governance

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/mailgun/mailgun-go/v4"
)

func LoadMailgunState(domain string) (*MailgunState, error) {
	ctx := context.Background()

	mailgunAPIKey := os.Getenv("MAILGUN_API_KEY")
	if mailgunAPIKey == "" {
		log.Fatalln("no $MAILGUN_API_KEY provided")
	}

	mg := mailgun.NewMailgun(domain, mailgunAPIKey)

	state := &MailgunState{}

	iter := mg.ListRoutes(nil)

	var routes []mailgun.Route
	for iter.Next(ctx, &routes) {
		for _, route := range routes {
			state.Routes = append(state.Routes, MailgunRoute{
				ID:          route.Id,
				Description: route.Description,
				Expression:  route.Expression,
				Actions:     route.Actions,
			})
		}
	}

	return state, nil
}

func (config *Config) DesiredMailgunState(domain string) *MailgunState {
	state := &MailgunState{}

	for _, team := range config.Teams {
		if len(team.RawMembers) == 0 {
			continue
		}

		route := MailgunRoute{
			Description: fmt.Sprintf("mailgun_route.routes[%q]", team.Name),
			Expression:  fmt.Sprintf("match_recipient(%q)", team.Name+"@"+domain),
		}

		for _, member := range team.Members(config) {
			if member.Email == "" {
				continue
			}

			route.Actions = append(
				route.Actions,
				fmt.Sprintf("forward(%q)", member.Email),
			)
		}

		route.Actions = append(
			route.Actions,
			"stop()",
		)

		state.Routes = append(state.Routes, route)
	}

	return state
}

type MailgunState struct {
	Routes []MailgunRoute
}

type MailgunRoute struct {
	ID          string
	Description string
	Expression  string
	Actions     []string
}
