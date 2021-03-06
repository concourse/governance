package governance

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Contributors map[string]Person
	Teams        map[string]Team
}

type Person struct {
	Name    string `yaml:"name"`
	GitHub  string `yaml:"github"`
	Discord string `yaml:"discord,omitempty"`
	Admin   bool   `yaml:"admin,omitempty"`
}

type Team struct {
	Name    string   `yaml:"name"`
	Purpose string   `yaml:"purpose"`
	Members []string `yaml:"members"`
	Repos   []string `yaml:"repos"`
}

func LoadConfig(tree fs.FS) (*Config, error) {
	personFiles, err := fs.ReadDir(tree, "contributors")
	if err != nil {
		return nil, err
	}

	contributors := map[string]Person{}
	for _, f := range personFiles {
		fn := filepath.Join("contributors", f.Name())

		file, err := tree.Open(fn)
		if err != nil {
			return nil, err
		}

		var person Person
		err = yaml.NewDecoder(file).Decode(&person)
		if err != nil {
			return nil, fmt.Errorf("decode %s: %w", fn, err)
		}

		contributors[strings.TrimSuffix(f.Name(), ".yml")] = person
	}

	teamFiles, err := fs.ReadDir(tree, "teams")
	if err != nil {
		return nil, err
	}

	teams := map[string]Team{}
	for _, f := range teamFiles {
		fn := filepath.Join("teams", f.Name())

		file, err := tree.Open(fn)
		if err != nil {
			return nil, err
		}

		var team Team
		err = yaml.NewDecoder(file).Decode(&team)
		if err != nil {
			return nil, fmt.Errorf("decode %s: %w", fn, err)
		}

		teams[strings.TrimSuffix(f.Name(), ".yml")] = team
	}

	return &Config{
		Contributors: contributors,
		Teams:        teams,
	}, nil
}

func (cfg Config) DesiredGitHubState() GitHubState {
	var state GitHubState

	for _, person := range cfg.Contributors {
		role := "MEMBER"
		if person.Admin {
			role = "ADMIN"
		}

		state.Members = append(state.Members, GitHubOrgMembership{
			Login: person.GitHub,
			Role:  role,
		})
	}

	for _, team := range cfg.Teams {
		ghTeam := GitHubTeam{
			Name:        team.Name,
			Description: strings.TrimSpace(strings.Join(strings.Split(team.Purpose, "\n"), " ")),
		}

		for _, m := range team.Members {
			role := "MEMBER"
			if cfg.Contributors[m].Admin {
				role = "MAINTAINER"
			}

			ghTeam.Members = append(ghTeam.Members, GitHubTeamMembership{
				Login: m,
				Role:  role,
			})
		}

		for _, r := range team.Repos {
			ghTeam.Repos = append(ghTeam.Repos, GitHubTeamRepoAccess{
				Name:       r,
				Permission: "MAINTAIN",
			})
		}

		state.Teams = append(state.Teams, ghTeam)
	}

	return state
}
