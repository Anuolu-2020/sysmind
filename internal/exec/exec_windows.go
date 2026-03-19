//go:build windows

package exec

import (
	"os/exec"
	"syscall"
)

// CREATE_NO_WINDOW prevents the creation of a console window when
// running a console application.
const CREATE_NO_WINDOW = 0x08000000

// hideWindow configures the command to run without creating a visible
// console window on Windows.
func hideWindow(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: CREATE_NO_WINDOW,
	}
}
