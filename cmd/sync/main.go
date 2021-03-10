package main

import (
	"log"

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
}
