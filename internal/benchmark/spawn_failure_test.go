package benchmark_test

import (
	"context"
	"testing"
	"time"

	"github.com/miklosn/cmdperf/internal/benchmark"
	"github.com/miklosn/cmdperf/internal/command"
)

// Spawn failures (command never started) must not be counted as successful
// runs or contribute garbage near-zero durations to timing statistics.
func TestSpawnFailureExcludedFromStats(t *testing.T) {
	testCommands := []*command.Command{
		{
			Raw:          "true",
			Shell:        "/nonexistent/shell",
			ShellOptions: []string{"-c"},
			Parallelism:  1,
		},
	}

	runner, err := benchmark.NewRunner(testCommands, benchmark.Options{
		Iterations:  5,
		Parallelism: 1,
		Timeout:     time.Second,
	})
	if err != nil {
		t.Fatalf("Failed to create runner: %v", err)
	}

	runner.Run(context.Background())

	stats := runner.Results[0]
	if stats.TotalRuns != 5 {
		t.Errorf("TotalRuns = %d, want 5", stats.TotalRuns)
	}
	if stats.ErrorCount != 5 {
		t.Errorf("ErrorCount = %d, want 5", stats.ErrorCount)
	}
	if stats.SuccessfulRuns != 0 {
		t.Errorf("SuccessfulRuns = %d, want 0 (spawn never succeeded)", stats.SuccessfulRuns)
	}
	if stats.Mean != 0 {
		t.Errorf("Mean = %v, want 0 (no valid timing samples)", stats.Mean)
	}
}
