package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/concourse/governance"
)

const organization = "concourse"

func main() {
	state, err := governance.LoadGitHubState(organization)
	if err != nil {
		log.Fatalln("failed to load GitHub state:", err)
	}

	config := state.ImpliedConfig()

	err = config.SyncMissing(".")
	if err != nil {
		log.Fatalln("failed sync missing config:", err)
	}

	tf, err := LoadTerraform()
	if err != nil {
		log.Fatalln("failed to load terraform:", err)
	}

	for _, member := range config.Contributors {
		tf.Import(
			fmt.Sprintf("github_membership.org[%q]", member.GitHub),
			organization+":"+member.GitHub,
		)
	}

	for _, repo := range config.Repos {
		tf.Import(
			fmt.Sprintf("github_repository.repos[%q]", repo.Name),
			repo.Name,
		)
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
			tf.Import(
				fmt.Sprintf("github_team_membership.members[%q]", team.Name+":"+member),
				strconv.Itoa(actualTeam.ID)+":"+member,
			)
		}

		for _, repo := range team.Repos {
			tf.Import(
				fmt.Sprintf("github_team_repository.repos[%q]", team.Name+":"+repo),
				strconv.Itoa(actualTeam.ID)+":"+repo,
			)
		}
	}
}
