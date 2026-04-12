//go:build windows

package command

import "os/exec"

func setSysProcAttr(cmd *exec.Cmd)  {}
func killProcessGroup(pid int)      {}
