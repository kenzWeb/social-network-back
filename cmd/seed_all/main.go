package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

type seederStep struct {
	name string
	path string
}

func main() {
	steps := []seederStep{
		{name: "users/follows", path: "./cmd/seeder"},
		{name: "stories", path: "./cmd/gen_stories"},
	}

	for _, step := range steps {
		if err := runSeeder(step); err != nil {
			log.Fatal(err)
		}
	}

	log.Println("All seeders completed successfully")
}

func runSeeder(step seederStep) error {
	log.Printf("Running %s seeder...", step.name)

	cmd := exec.Command("go", "run", step.path)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s seeder failed: %w", step.name, err)
	}

	return nil
}