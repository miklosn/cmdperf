//go:build !windows

package main

const (
	defaultShell    = "/bin/sh"
	defaultShellOpt = "-c"
)
