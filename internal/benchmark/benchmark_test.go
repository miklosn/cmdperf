package benchmark_test

import (
	"context"
	"testing"
	"time"

	"github.com/miklosn/cmdperf/internal/benchmark"
	"github.com/miklosn/cmdperf/internal/command"
)

func TestRunner(t *testing.T) {
	testCommands := []*command.Command{
		{
			Raw:          "echo test1",
			Shell:        "/bin/sh",
			ShellOptions: []string{"-c"},
			Parallelism:  2,
		},
		{
			Raw:          "echo test2",
			Shell:        "/bin/sh",
			ShellOptions: []string{"-c"},
			Parallelism:  2,
		},
	}

	benchmarkOptions := benchmark.Options{
		Iterations:  10,
		Parallelism: 2,
		Timeout:     time.Second,
	}

	runner, err := benchmark.NewRunner(testCommands, benchmarkOptions)
	if err != nil {
		t.Fatalf("Failed to create runner: %v", err)
	}

	progressCalled := false
	completionCalled := false
	runner.SetProgressCallback(func(stats []*benchmark.CommandStats, complete bool) {
		if complete {
			completionCalled = true
		} else {
			progressCalled = true
		}
	})

	runner.Run(context.Background())
	if !progressCalled {
		t.Error("Progress callback was never called")
	}
	if !completionCalled {
		t.Error("Completion callback was never called")
	}

	if len(runner.Results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(runner.Results))
	}

	firstCmdStats := runner.Results[0]
	if firstCmdStats.Command != testCommands[0] {
		t.Error("Command reference mismatch in results")
	}
	if firstCmdStats.TotalRuns != benchmarkOptions.Iterations {
		t.Errorf("Expected %d total runs, got %d", benchmarkOptions.Iterations, firstCmdStats.TotalRuns)
	}
	if firstCmdStats.Min <= 0 {
		t.Errorf("Expected positive Min duration, got %v", firstCmdStats.Min)
	}
	if firstCmdStats.Max <= 0 {
		t.Errorf("Expected positive Max duration, got %v", firstCmdStats.Max)
	}
	if firstCmdStats.Mean <= 0 {
		t.Errorf("Expected positive Mean duration, got %v", firstCmdStats.Mean)
	}
}
