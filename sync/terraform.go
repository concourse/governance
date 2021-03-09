package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

type Terraform struct {
	state map[string]bool
}

func LoadTerraform() (*Terraform, error) {
	out := new(bytes.Buffer)

	list := exec.Command("terraform", "state", "list")
	list.Stdout = out
	list.Stderr = os.Stderr

	err := list.Run()
	if err != nil {
		return nil, fmt.Errorf("terraform state list: %w", err)
	}

	state := map[string]bool{}
	for _, resource := range strings.Split(out.String(), "\n") {
		if resource == "" {
			continue
		}

		state[resource] = true
	}

	return &Terraform{state}, nil
}

func (tf Terraform) Import(resource, id string) {
	if tf.state[resource] {
		// already exists
		return
	}

	cmd := exec.Command("terraform", "import", resource, id)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("\x1b[33m==== EXEC %s\x1b[0m\n", strings.Join(cmd.Args, " "))

	err := cmd.Run()
	if err != nil {
		log.Fatalln("failed to import:", err)
	}
}
