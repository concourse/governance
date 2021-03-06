package governance

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

type GitHubState struct {
	Organization string

	Members []GitHubOrgMembership
	Teams   []GitHubTeam
	Repos   []GitHubRepo
}

func (state GitHubState) Team(name string) (GitHubTeam, bool) {
	for _, team := range state.Teams {
		if team.Name == name {
			return team, true
		}
	}

	return GitHubTeam{}, false
}

type GitHubOrgMembership struct {
	Login string
	Role  string
}

type GitHubTeam struct {
	Name        string
	Description string
	Members     []GitHubTeamMembership
	Repos       []GitHubTeamRepoAccess
}

type GitHubTeamMembership struct {
	Login string
	Role  string
}

type GitHubTeamRepoAccess struct {
	Name       string
	Permission string
}

type GitHubRepo struct {
	Name                string
	DirectCollaborators []GitHubRepoCollaborator
}

type GitHubRepoCollaborator struct {
	Login      string
	Permission string
}

func LoadGitHubState(orgName string) (*GitHubState, error) {
	ctx := context.Background()

	var tc *http.Client
	var githubToken = os.Getenv("GITHUB_TOKEN")

	if githubToken != "" {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: githubToken},
		)

		tc = oauth2.NewClient(ctx, ts)
	}

	client := githubv4.NewClient(tc)

	org := &GitHubState{
		Organization: orgName,
	}

	err := org.LoadMembers(ctx, client)
	if err != nil {
		return nil, err
	}

	err = org.LoadTeams(ctx, client)
	if err != nil {
		return nil, err
	}

	err = org.LoadRepos(ctx, client)
	if err != nil {
		return nil, err
	}

	return org, nil
}

func (state *GitHubState) LoadMembers(ctx context.Context, client *githubv4.Client) error {
	args := map[string]interface{}{
		"org":   githubv4.String(state.Organization),
		"limit": githubv4.Int(100),
		"after": (*githubv4.String)(nil),
	}

	for {
		var membersQ struct {
			Organization struct {
				Members struct {
					Edges []struct {
						Role string
						Node struct {
							Login string
						}
					}
					PageInfo struct {
						EndCursor   githubv4.String
						HasNextPage bool
					}
				} `graphql:"membersWithRole(first: $limit, after: $after)"`
			} `graphql:"organization(login: $org)"`
		}
		err := client.Query(ctx, &membersQ, args)
		if err != nil {
			return fmt.Errorf("list members: %w", err)
		}

		for _, edge := range membersQ.Organization.Members.Edges {
			state.Members = append(state.Members, GitHubOrgMembership{
				Login: edge.Node.Login,
				Role:  edge.Role,
			})
		}

		if !membersQ.Organization.Members.PageInfo.HasNextPage {
			break
		}

		args["after"] = githubv4.NewString(membersQ.Organization.Members.PageInfo.EndCursor)
	}

	return nil
}

func (state *GitHubState) LoadTeams(ctx context.Context, client *githubv4.Client) error {
	var teamsQ struct {
		Organization struct {
			Teams struct {
				Nodes []struct {
					Name        string
					Description string

					Members struct {
						Edges []struct {
							Role string
							Node struct {
								Login string
							}
						}
					} `graphql:"members(first: 100)"` // 100 ought to be enough

					Repositories struct {
						Edges []struct {
							Permission string
							Node       struct {
								Name string
							}
						}
					} `graphql:"repositories(first: 100)"` // 100 ought to be enough
				}
			} `graphql:"teams(first: 100)"` // 100 ought to be enough
		} `graphql:"organization(login: $org)"`
	}
	err := client.Query(ctx, &teamsQ, map[string]interface{}{
		"org": githubv4.String(state.Organization),
	})
	if err != nil {
		return fmt.Errorf("list teams: %w", err)
	}

	for _, node := range teamsQ.Organization.Teams.Nodes {
		team := GitHubTeam{
			Name:        node.Name,
			Description: node.Description,
		}

		for _, edge := range node.Members.Edges {
			team.Members = append(team.Members, GitHubTeamMembership{
				Login: edge.Node.Login,
				Role:  edge.Role,
			})
		}

		for _, edge := range node.Repositories.Edges {
			team.Repos = append(team.Repos, GitHubTeamRepoAccess{
				Name:       edge.Node.Name,
				Permission: edge.Permission,
			})
		}

		state.Teams = append(state.Teams, team)
	}

	return nil
}

func (state *GitHubState) LoadRepos(ctx context.Context, client *githubv4.Client) error {
	var reposQ struct {
		Organization struct {
			Repositories struct {
				Nodes []struct {
					Name string

					Collaborators struct {
						Edges []struct {
							Permission string
							Node       struct {
								Login string
							}
						}
					} `graphql:"collaborators(first: 100, affiliation: DIRECT)"` // 100 ought to be enough
				}
			} `graphql:"repositories(first: 100)"` // 100 ought to be enough
		} `graphql:"organization(login: $org)"`
	}
	err := client.Query(ctx, &reposQ, map[string]interface{}{
		"org": githubv4.String(state.Organization),
	})
	if err != nil {
		if strings.Contains(err.Error(), "Must have push access to view repository collaborators.") {
			// swallow error caused by archived repos; reposQ will still be populated
			// with the response
		} else {
			return err
		}
	}

	for _, node := range reposQ.Organization.Repositories.Nodes {
		repo := GitHubRepo{
			Name: node.Name,
		}

		for _, edge := range node.Collaborators.Edges {
			repo.DirectCollaborators = append(repo.DirectCollaborators, GitHubRepoCollaborator{
				Login:      edge.Node.Login,
				Permission: edge.Permission,
			})
		}

		state.Repos = append(state.Repos, repo)
	}

	return nil
}
