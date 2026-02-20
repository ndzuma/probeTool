//go:build !windows

package cmd

import (
	"os/exec"
	"syscall"
)

func startDaemon(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}
}
