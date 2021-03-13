package governance

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

type OrgRole string

const OrgRoleMember OrgRole = "MEMBER"
const OrgRoleAdmin OrgRole = "ADMIN"

type TeamRole string

const TeamRoleMember TeamRole = "MEMBER"
const TeamRoleMaintainer TeamRole = "MAINTAINER"

type RepoPermission string

const RepoPermissionMaintain RepoPermission = "MAINTAIN"

type GitHubState struct {
	Organization string

	Members []GitHubOrgMember
	Teams   []GitHubTeam
	Repos   []GitHubRepo
}

func (state GitHubState) Member(login string) (GitHubOrgMember, bool) {
	for _, member := range state.Members {
		if member.Login == login {
			return member, true
		}
	}

	return GitHubOrgMember{}, false
}

func (state GitHubState) Team(name string) (GitHubTeam, bool) {
	for _, team := range state.Teams {
		if team.Name == name {
			return team, true
		}
	}

	return GitHubTeam{}, false
}

func (state GitHubState) Repo(name string) (GitHubRepo, bool) {
	for _, repo := range state.Repos {
		if repo.Name == name {
			return repo, true
		}
	}

	return GitHubRepo{}, false
}

type GitHubOrgMember struct {
	Name  string
	Login string
	Role  OrgRole
}

type GitHubTeam struct {
	ID          int
	Name        string
	Description string
	Members     []GitHubTeamMember
	Repos       []GitHubTeamRepoAccess
}

func (team GitHubTeam) Member(login string) (GitHubTeamMember, bool) {
	for _, member := range team.Members {
		if member.Login == login {
			return member, true
		}
	}

	return GitHubTeamMember{}, false
}

func (team GitHubTeam) Repo(name string) (GitHubTeamRepoAccess, bool) {
	for _, repo := range team.Repos {
		if repo.Name == name {
			return repo, true
		}
	}

	return GitHubTeamRepoAccess{}, false
}

type GitHubTeamMember struct {
	Login string
	Role  TeamRole
}

type GitHubTeamRepoAccess struct {
	Name       string
	Permission RepoPermission
}

type GitHubRepo struct {
	Name                string
	Description         string
	IsPrivate           bool
	Topics              []string
	HomepageURL         string
	HasIssues           bool
	HasWiki             bool
	HasProjects         bool
	DirectCollaborators []GitHubRepoCollaborator
}

func (repo GitHubRepo) Collaborator(login string) (GitHubRepoCollaborator, bool) {
	for _, collaborator := range repo.DirectCollaborators {
		if collaborator.Login == login {
			return collaborator, true
		}
	}

	return GitHubRepoCollaborator{}, false
}

type GitHubRepoCollaborator struct {
	Login      string
	Permission string
}

func LoadGitHubState(orgName string) (*GitHubState, error) {
	ctx := context.Background()

	githubToken := os.Getenv("GITHUB_TOKEN")
	if githubToken == "" {
		return nil, fmt.Errorf("no $GITHUB_TOKEN provided")
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubToken},
	)

	client := githubv4.NewClient(oauth2.NewClient(ctx, ts))

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

func (state *GitHubState) ImpliedConfig() Config {
	config := Config{
		Contributors: map[string]Person{},
		Teams:        map[string]Team{},
		Repos:        map[string]Repo{},
	}

	for _, t := range state.Teams {
		team := Team{
			Name:    t.Name,
			Purpose: t.Description,
		}

		for _, member := range t.Members {
			team.Members = append(team.Members, member.Login)
		}

		for _, repo := range t.Repos {
			if repo.Permission != RepoPermissionMaintain {
				// this model only sets teams up with Maintain permission, so skip
				// anything else
				continue
			}

			team.Repos = append(team.Repos, repo.Name)
		}

		config.Teams[team.Name] = team
	}

	contributorRepos := map[string]map[string]string{}
	for _, repo := range state.Repos {
		config.Repos[repo.Name] = Repo{
			Name:        repo.Name,
			Description: repo.Description,
			Private:     repo.IsPrivate,
			Topics:      repo.Topics,
			HomepageURL: repo.HomepageURL,
			HasIssues:   repo.HasIssues,
			HasProjects: repo.HasProjects,
			HasWiki:     repo.HasWiki,
		}

		for _, collaborator := range repo.DirectCollaborators {
			repos, found := contributorRepos[collaborator.Login]
			if !found {
				repos = map[string]string{}
				contributorRepos[collaborator.Login] = repos
			}

			repos[repo.Name] = permission4to3(collaborator.Permission)
		}
	}

	for _, member := range state.Members {
		config.Contributors[member.Login] = Person{
			Name:   member.Name,
			GitHub: member.Login,
			Repos:  contributorRepos[member.Login],
		}
	}

	return config
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
							Name  string
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
			role := OrgRole(edge.Role)
			if role == OrgRoleAdmin {
				continue
			}

			state.Members = append(state.Members, GitHubOrgMember{
				Name:  edge.Node.Name,
				Login: edge.Node.Login,
				Role:  role,
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
					DatabaseId  int
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
			ID:          node.DatabaseId,
			Name:        node.Name,
			Description: node.Description,
		}

		for _, edge := range node.Members.Edges {
			team.Members = append(team.Members, GitHubTeamMember{
				Login: edge.Node.Login,
				Role:  TeamRole(edge.Role),
			})
		}

		for _, edge := range node.Repositories.Edges {
			team.Repos = append(team.Repos, GitHubTeamRepoAccess{
				Name:       edge.Node.Name,
				Permission: RepoPermission(edge.Permission),
			})
		}

		state.Teams = append(state.Teams, team)
	}

	return nil
}

func (state *GitHubState) LoadRepos(ctx context.Context, client *githubv4.Client) error {
	args := map[string]interface{}{
		"org":   githubv4.String(state.Organization),
		"limit": githubv4.Int(100),
		"after": (*githubv4.String)(nil),
	}

	for {
		var reposQ struct {
			Organization struct {
				Repositories struct {
					Nodes []struct {
						Name string

						Description string

						Topics struct {
							Nodes []struct {
								Topic struct {
									Name string
								}
							}
						} `graphql:"repositoryTopics(first: 10)"` // 10 ought to be enough

						HomepageURL string

						IsArchived bool
						IsPrivate  bool

						HasIssuesEnabled   bool
						HasProjectsEnabled bool
						HasWikiEnabled     bool

						Collaborators struct {
							Edges []struct {
								Permission string
								Node       struct {
									Login string
								}
							}
						} `graphql:"collaborators(first: 100, affiliation: DIRECT)"` // 100 ought to be enough
					}

					PageInfo struct {
						EndCursor   githubv4.String
						HasNextPage bool
					}
				} `graphql:"repositories(first: $limit, after: $after)"`
			} `graphql:"organization(login: $org)"`
		}
		err := client.Query(ctx, &reposQ, args)
		if err != nil {
			if strings.Contains(err.Error(), "Must have push access to view repository collaborators.") {
				// swallow error caused by archived repos; reposQ will still be populated
				// with the response
			} else {
				return err
			}
		}

		for _, node := range reposQ.Organization.Repositories.Nodes {
			if node.IsArchived {
				continue
			}

			repo := GitHubRepo{
				Name:        node.Name,
				Description: node.Description,
				IsPrivate:   node.IsPrivate,
				HomepageURL: node.HomepageURL,
				HasIssues:   node.HasIssuesEnabled,
				HasProjects: node.HasProjectsEnabled,
				HasWiki:     node.HasWikiEnabled,
			}

			for _, node := range node.Topics.Nodes {
				repo.Topics = append(repo.Topics, node.Topic.Name)
			}

			for _, edge := range node.Collaborators.Edges {
				repo.DirectCollaborators = append(repo.DirectCollaborators, GitHubRepoCollaborator{
					Login:      edge.Node.Login,
					Permission: edge.Permission,
				})
			}

			state.Repos = append(state.Repos, repo)
		}

		if !reposQ.Organization.Repositories.PageInfo.HasNextPage {
			break
		}

		args["after"] = githubv4.NewString(reposQ.Organization.Repositories.PageInfo.EndCursor)
	}

	return nil
}
