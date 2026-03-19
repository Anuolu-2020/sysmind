// Package exec provides platform-aware command execution utilities.
// On Windows, it hides console windows that would otherwise appear when
// running external commands like netstat, docker, etc.
package exec

import (
	"os/exec"
)

// Cmd is an alias to os/exec.Cmd for convenience.
type Cmd = exec.Cmd

// Command creates an exec.Cmd with platform-appropriate settings.
// On Windows, this configures the command to run without creating
// a visible console window (preventing "virus-like" behavior).
// On other platforms, it behaves identically to exec.Command.
func Command(name string, args ...string) *Cmd {
	cmd := exec.Command(name, args...)
	hideWindow(cmd)
	return cmd
}
