package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestCSVWriter(t *testing.T) {
	stats := createTestStats()

	var buf bytes.Buffer
	writer := &CSVWriter{}
	err := writer.Write(&buf, stats)

	if err != nil {
		t.Fatalf("Failed to write CSV: %v", err)
	}

	output := buf.String()

	expectedHeaders := "Command,TotalRuns,SuccessfulRuns,ErrorCount,NonZeroExitCodes,Min (ns),Max (ns),Mean (ns),Median (ns),StdDev (ns),Throughput (/s)"
	if !strings.Contains(output, expectedHeaders) {
		t.Errorf("CSV output missing expected headers")
	}

	if !strings.Contains(output, "echo hello") {
		t.Errorf("CSV output missing command data")
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 3 {
		t.Errorf("Expected 3 lines (header + 2 commands), got %d", len(lines))
	}
}
