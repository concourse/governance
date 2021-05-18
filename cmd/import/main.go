package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/concourse/governance"
	"github.com/google/go-github/v35/github"
	"github.com/mailgun/mailgun-go/v4"
	"golang.org/x/oauth2"
)

const organization = "concourse"
const domain = "concourse-ci.org"

func main() {
	config, err := governance.LoadConfig(os.DirFS("."))
	if err != nil {
		log.Fatalln("failed to load config:", err)
	}

	state, err := governance.LoadGitHubState(organization)
	if err != nil {
		log.Fatalln("failed to load GitHub state:", err)
	}

	mailgunAPIKey := os.Getenv("MAILGUN_API_KEY")
	if mailgunAPIKey == "" {
		log.Fatalln("no $MAILGUN_API_KEY provided")
	}

	mg := mailgun.NewMailgun(domain, mailgunAPIKey)

	tf, err := LoadTerraform()
	if err != nil {
		log.Fatalln("failed to load terraform:", err)
	}

	ctx := context.Background()

	// v3 client is necessary for getting deploy key numeric IDs, since that's
	// used as the ID in the github terraform provider. :(
	v3client := newv3Client(ctx)

	for _, member := range config.Contributors {
		_, found := state.Member(member.GitHub)
		if !found {
			continue
		}

		tf.Import(
			fmt.Sprintf("github_membership.contributors[%q]", member.GitHub),
			organization+":"+member.GitHub,
		)

		for repo := range member.Repos {
			actualRepo, found := state.Repo(repo)
			if !found {
				continue
			}

			_, found = actualRepo.Collaborator(member.GitHub)
			if !found {
				continue
			}

			tf.Import(
				fmt.Sprintf("github_repository_collaborator.collaborators[%q]", repo+":"+member.GitHub),
				repo+":"+member.GitHub,
			)
		}
	}

	for _, repo := range config.Repos {
		_, found := state.Repo(repo.Name)
		if !found {
			continue
		}

		tf.Import(
			fmt.Sprintf("github_repository.repos[%q]", repo.Name),
			repo.Name,
		)

		for _, protection := range repo.BranchProtection {
			tf.Import(
				fmt.Sprintf("github_branch_protection.branch_protections[%q]", repo.Name+":"+protection.Pattern),
				repo.Name+":"+protection.Pattern,
			)
		}

		for _, label := range repo.Labels {
			tf.Import(
				fmt.Sprintf("github_issue_label.labels[%q]", repo.Name+":"+label.Name),
				repo.Name+":"+label.Name,
			)
		}

		v3keys, _, err := v3client.Repositories.ListKeys(ctx, organization, repo.Name, &github.ListOptions{})
		if err != nil {
			log.Fatalf("failed to list deploy keys for repo %s: %s", repo.Name, err)
			return
		}

		for _, key := range repo.DeployKeys {
			var keyId int64
			for _, k := range v3keys {
				if k.GetTitle() == key.Title {
					keyId = k.GetID()
					break
				}
			}
			if keyId == 0 {
				log.Println("did not find key id for title:", key.Title)
				continue
			}

			tf.Import(
				fmt.Sprintf("github_repository_deploy_key.keys[%q]", repo.Name+":"+key.Title),
				fmt.Sprintf("%s:%d", repo.Name, keyId),
			)
		}
	}

	for _, team := range config.Teams {
		actualTeam, found := state.Team(team.Name)
		if !found {
			continue
		}

		tf.Import(
			fmt.Sprintf("github_team.teams[%q]", team.Name),
			strconv.Itoa(actualTeam.ID),
		)

		for _, person := range team.Members(config) {
			_, found := actualTeam.Member(person.GitHub)
			if !found {
				continue
			}

			tf.Import(
				fmt.Sprintf("github_team_membership.members[%q]", team.Name+":"+person.GitHub),
				strconv.Itoa(actualTeam.ID)+":"+person.GitHub,
			)
		}

		for _, repo := range team.Repos {
			_, found := actualTeam.Repo(repo)
			if !found {
				continue
			}

			tf.Import(
				fmt.Sprintf("github_team_repository.repos[%q]", team.Name+":"+repo),
				strconv.Itoa(actualTeam.ID)+":"+repo,
			)
		}
	}

	iter := mg.ListRoutes(nil)

	var routes []mailgun.Route
	for iter.Next(ctx, &routes) {
		for _, route := range routes {
			// sneaky way of making it easy to import routes back into whichever
			// resource created them
			tf.Import(route.Description, fmt.Sprintf("us:%s", route.Id))
		}
	}
}

func newv3Client(ctx context.Context) *github.Client {
	githubToken := os.Getenv("GITHUB_TOKEN")
	if githubToken == "" {
		log.Fatalln("no github token provided... somehow")
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc)
}
