//go:build !windows
package adb

import "os/exec"

func PrepareCmd(cmd *exec.Cmd) {}
