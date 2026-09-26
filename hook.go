package main

import (
	"fmt"
	"os"
	"os/exec"
)

// runHook execs the user's command with the flushed file appended as the
// final argument. Deliberately a plain exec, not a shell: the hook is a
// program, and programs don't need quoting rules.
func runHook(cmd, file string) {
	c := exec.Command(cmd, file)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "hook:", err)
	}
}
