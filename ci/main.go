// ci is the entire CI pipeline as a Go program — the house rule: workflows
// invoke exactly one thing, and that thing is Go, not bash.
package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	steps := [][]string{
		{"go", "vet", "./..."},
		{"go", "test", "./..."},
		{"go", "build", "./..."},
	}
	for _, step := range steps {
		fmt.Printf("--> %v\n", step)
		cmd := exec.Command(step[0], step[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "ci: %v failed: %v\n", step, err)
			os.Exit(1)
		}
	}
	fmt.Println("ci: green")
}
