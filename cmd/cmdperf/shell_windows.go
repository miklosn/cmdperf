//go:build windows

package main

import "os"

var (
	defaultShell    = comspec()
	defaultShellOpt = "/c"
)

func comspec() string {
	if shell := os.Getenv("COMSPEC"); shell != "" {
		return shell
	}
	return "cmd.exe"
}
