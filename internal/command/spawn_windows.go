//go:build windows

package command

import (
	"os/exec"
	"strconv"
)

func setSysProcAttr(cmd *exec.Cmd) {}

// killProcessGroup kills the process and its whole tree. Without this,
// timed-out commands (e.g. anything blocking on stdin under cmd.exe) leave
// children running and the worker blocks forever on Wait.
func killProcessGroup(pid int) {
	_ = exec.Command("taskkill", "/F", "/T", "/PID", strconv.Itoa(pid)).Run()
}
