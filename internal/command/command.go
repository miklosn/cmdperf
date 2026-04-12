package command

import (
	"context"
	"os/exec"
	"sync"
	"time"
)

// Command represents a shell command to be benchmarked
type Command struct {
	Raw          string
	Shell        string
	ShellOptions []string
	Timeout      time.Duration
	Parallelism  int

	// Cached shell options to avoid repeated allocations
	cachedShellOptions []string

	// DirectExec indicates whether to execute the command directly without a shell
	DirectExec bool

	Command string
	Args    []string
}

// Result represents the result of a single command execution
type Result struct {
	Command          *Command
	StartTime        time.Time
	Duration         time.Duration
	ExitCode         int
	Error            error
	TimedOut         bool
	ContextCancelled bool // New field to track context cancellation
}

// Object pool for Result objects to reduce allocations
var resultPool = sync.Pool{
	New: func() interface{} {
		return &Result{}
	},
}

// Execute runs the command once and returns the result
func (c *Command) Execute(ctx context.Context) *Result {
	// Get a result from the pool instead of allocating a new one
	result := resultPool.Get().(*Result)

	// Reset the result fields
	result.Command = c
	result.StartTime = time.Now()
	result.Duration = 0
	result.ExitCode = 0
	result.Error = nil
	result.TimedOut = false
	result.ContextCancelled = false // Reset the new field

	// Check if context is already cancelled
	if ctx.Err() != nil {
		result.Error = ctx.Err()
		result.ContextCancelled = true
		result.Duration = time.Since(result.StartTime)
		return result
	}

	// Create a context with timeout if timeout is set
	var execCtx context.Context
	var cancel context.CancelFunc
	if c.Timeout > 0 {
		execCtx, cancel = context.WithTimeout(ctx, c.Timeout)
		defer cancel()
	} else {
		execCtx = ctx
		cancel = func() {} // No-op cancel function
	}

	// Create a command based on execution mode
	var cmd *exec.Cmd
	if c.DirectExec {
		// Direct execution mode
		cmd = exec.CommandContext(execCtx, c.Command, c.Args...)
		setSysProcAttr(cmd)
	} else {
		// Shell execution mode - use cached options if available
		if c.cachedShellOptions == nil {
			c.cachedShellOptions = make([]string, len(c.ShellOptions)+1)
			copy(c.cachedShellOptions, c.ShellOptions)
			c.cachedShellOptions[len(c.ShellOptions)] = c.Raw
		}
		cmd = exec.CommandContext(execCtx, c.Shell, c.cachedShellOptions...)
		setSysProcAttr(cmd)
	}

	// Execute and capture timing
	startTime := time.Now()
	doneCh := make(chan struct{})
	if err := cmd.Start(); err != nil {
		result.Error = err
		result.ExitCode = -1
		result.Duration = time.Since(startTime)
		return result
	}
	go func() {
		select {
		case <-execCtx.Done():
			if cmd.Process != nil {
				killProcessGroup(cmd.Process.Pid)
			}
		case <-doneCh:
		}
	}()
	err := cmd.Wait()
	close(doneCh)
	endTime := time.Now()
	result.Duration = endTime.Sub(startTime)

	// Handle execution results
	if err != nil {
		result.Error = err
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = -1
		}

		if execCtx.Err() == context.DeadlineExceeded {
			result.TimedOut = true
		}

		// Check if parent context was cancelled (duration elapsed)
		if ctx.Err() != nil {
			result.ContextCancelled = true
		}
	}

	return result
}

// ReleaseResult returns a Result to the pool for reuse
// This should be called when the Result is no longer needed
func ReleaseResult(r *Result) {
	if r != nil {
		resultPool.Put(r)
	}
}
