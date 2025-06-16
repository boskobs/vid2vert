//go:build windows
// +build windows

package helpers

import (
	"os/exec"
	"syscall"
)

// Command creates an exec.Cmd for the given command and arguments.
// On Windows, it sets the CreationFlags to 0x08000000 to hide the console.
func Command(command string, args ...string) *exec.Cmd {
	cmd := exec.Command(command, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x08000000}
	return cmd
}
