package command_test

import (
	"context"
	"testing"
	"time"

	"github.com/miklosn/cmdperf/internal/command"
)

func TestSimpleCommand(t *testing.T) {
	cmd := &command.Command{
		Raw:          "echo test",
		Shell:        "/bin/sh",
		ShellOptions: []string{"-c"},
	}

	result := cmd.Execute(context.Background())

	if result.Error != nil {
		t.Errorf("Expected no error, got: %v", result.Error)
	}
	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got: %d", result.ExitCode)
	}
	if result.Duration <= 0 {
		t.Errorf("Expected positive duration, got: %v", result.Duration)
	}
}

func TestFailingCommand(t *testing.T) {
	cmd := &command.Command{
		Raw:          "exit 1",
		Shell:        "/bin/sh",
		ShellOptions: []string{"-c"},
	}

	result := cmd.Execute(context.Background())

	if result.Error == nil {
		t.Error("Expected an error, got nil")
	}
	if result.ExitCode != 1 {
		t.Errorf("Expected exit code 1, got: %d", result.ExitCode)
	}
}

func TestCommandTimeout(t *testing.T) {
	cmd := &command.Command{
		Raw:          "sleep 1",
		Shell:        "/bin/sh",
		ShellOptions: []string{"-c"},
		Timeout:      time.Millisecond * 100, // Very short timeout
	}

	ctx, cancel := context.WithTimeout(context.Background(), cmd.Timeout)
	defer cancel()

	result := cmd.Execute(ctx)

	if !result.TimedOut {
		t.Error("Expected command to time out, but it didn't")
	}
	if result.Error == nil {
		t.Error("Expected an error due to timeout, got nil")
	}
}
