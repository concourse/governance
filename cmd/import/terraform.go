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

func LoadTerraform() (Terraform, error) {
	out := new(bytes.Buffer)

	tf := Terraform{
		state: map[string]bool{},
	}

	_, err := os.Stat("terraform.tfstate")
	if err != nil {
		if os.IsNotExist(err) {
			return tf, nil
		}

		return Terraform{}, err
	}

	list := exec.Command("terraform", "state", "list")
	list.Stdout = out
	list.Stderr = os.Stderr

	err = list.Run()
	if err != nil {
		return Terraform{}, fmt.Errorf("terraform state list: %w", err)
	}

	for _, resource := range strings.Split(out.String(), "\n") {
		if resource == "" {
			continue
		}

		tf.state[resource] = true
	}

	return tf, nil
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
