//go:build windows
package adb

import (
	"os/exec"
	"syscall"
)

func PrepareCmd(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}
}
