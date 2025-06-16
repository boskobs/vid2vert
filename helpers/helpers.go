//go:build !windows
// +build !windows

package helpers

import (
	"os/exec"
)

// Command creates an exec.Cmd for the given command and arguments.
func Command(command string, args ...string) *exec.Cmd {
	cmd := exec.Command(command, args...)
	return cmd
}
