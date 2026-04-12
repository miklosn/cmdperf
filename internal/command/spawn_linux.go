//go:build linux

package command

import (
	"os/exec"
	"syscall"
)

// vfork requires Credential, Ptrace, Foreground, and AmbientCaps to be
// unset. cmdperf never sets these; adding any will silently break vfork.
func setSysProcAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid:    true,
		Cloneflags: syscall.CLONE_VFORK,
	}
}

func killProcessGroup(pid int) {
	_ = syscall.Kill(-pid, syscall.SIGKILL)
}
