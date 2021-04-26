package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/concourse/governance"
)

const organization = "concourse"

func main() {
	config, err := governance.LoadConfig(os.DirFS("."))
	if err != nil {
		log.Fatalln("failed to load config:", err)
	}

	state, err := governance.LoadGitHubState(organization)
	if err != nil {
		log.Fatalln("failed to load GitHub state:", err)
	}

	tf, err := LoadTerraform()
	if err != nil {
		log.Fatalln("failed to load terraform:", err)
	}

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

		for _, member := range team.Members {
			_, found := actualTeam.Member(member)
			if !found {
				continue
			}

			tf.Import(
				fmt.Sprintf("github_team_membership.members[%q]", team.Name+":"+member),
				strconv.Itoa(actualTeam.ID)+":"+member,
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
}
