package output

import (
	"testing"
	"time"

	"github.com/miklosn/cmdperf/internal/benchmark"
	"github.com/miklosn/cmdperf/internal/command"
)

func TestGetWriter(t *testing.T) {
	tests := []struct {
		format      string
		expectError bool
		writerType  string
	}{
		{"csv", false, "*output.CSVWriter"},
		{"markdown", false, "*output.MarkdownWriter"},
		{"terminal", false, "*output.TerminalWriter"},
		{"invalid", true, ""},
	}

	for _, test := range tests {
		writer, err := GetWriter(test.format)

		if test.expectError {
			if err == nil {
				t.Errorf("GetWriter(%s) expected error, got nil", test.format)
			}
		} else {
			if err != nil {
				t.Errorf("GetWriter(%s) unexpected error: %v", test.format, err)
			}

			typeName := getTypeName(writer)
			if typeName != test.writerType {
				t.Errorf("GetWriter(%s) = %s, expected %s",
					test.format, typeName, test.writerType)
			}
		}
	}
}

func getTypeName(i interface{}) string {
	if i == nil {
		return "<nil>"
	}
	return "*output." + getType(i)
}

func getType(i interface{}) string {
	switch i.(type) {
	case *CSVWriter:
		return "CSVWriter"
	case *MarkdownWriter:
		return "MarkdownWriter"
	case *TerminalWriter:
		return "TerminalWriter"
	default:
		return "Unknown"
	}
}

func createTestStats() []*benchmark.CommandStats {
	cmd1 := &command.Command{
		Raw:          "echo hello",
		Shell:        "/bin/sh",
		ShellOptions: []string{"-c"},
		Timeout:      time.Second,
		Parallelism:  1,
	}

	cmd2 := &command.Command{
		Raw:          "sleep 0.1",
		Shell:        "/bin/sh",
		ShellOptions: []string{"-c"},
		Timeout:      time.Second,
		Parallelism:  1,
	}

	stats1 := &benchmark.CommandStats{
		Command:        cmd1,
		TotalRuns:      100,
		SuccessfulRuns: 100,
		ErrorCount:     0,
		Min:            time.Millisecond,
		Max:            5 * time.Millisecond,
		Mean:           2 * time.Millisecond,
		Median:         2 * time.Millisecond,
		StdDev:         time.Millisecond,
		Throughput:     500,
		ExitCodes:      map[int]int{0: 100},
	}

	stats2 := &benchmark.CommandStats{
		Command:        cmd2,
		TotalRuns:      100,
		SuccessfulRuns: 95,
		ErrorCount:     5,
		Min:            100 * time.Millisecond,
		Max:            120 * time.Millisecond,
		Mean:           110 * time.Millisecond,
		Median:         110 * time.Millisecond,
		StdDev:         5 * time.Millisecond,
		Throughput:     9,
		ExitCodes:      map[int]int{0: 95, 1: 5},
	}

	return []*benchmark.CommandStats{stats1, stats2}
}
