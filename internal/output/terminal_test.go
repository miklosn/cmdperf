package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestTerminalWriter(t *testing.T) {
	stats := createTestStats()

	var buf bytes.Buffer
	writer := &TerminalWriter{}
	err := writer.Write(&buf, stats)

	if err != nil {
		t.Fatalf("Failed to write Terminal output: %v", err)
	}

	output := buf.String()

	expectedContent := []string{
		"✨ cmdperf - Command Performance Benchmarking ✨",
		"Command:",
		"echo hello",
		"sleep 0.1",
		"Comparison:",
	}

	for _, content := range expectedContent {
		if !strings.Contains(output, content) {
			t.Errorf("Terminal output missing expected content: %s", content)
		}
	}

	if !strings.Contains(output, "slower than") {
		t.Errorf("Terminal output missing comparison data")
	}
}
