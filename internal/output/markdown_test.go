package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestMarkdownWriter(t *testing.T) {
	stats := createTestStats()

	var buf bytes.Buffer
	writer := &MarkdownWriter{}
	err := writer.Write(&buf, stats)

	if err != nil {
		t.Fatalf("Failed to write Markdown: %v", err)
	}

	output := buf.String()

	expectedSections := []string{
		"# ✨ cmdperf - Command Performance Benchmarking ✨",
		"## Benchmark Parameters",
		"## Summary",
		"| Command | Runs | Mean ± StdDev | Min | Max | Throughput | Errors |",
		"## Command Parameters",
		"## Comparison",
	}

	for _, section := range expectedSections {
		if !strings.Contains(output, section) {
			t.Errorf("Markdown output missing expected section: %s", section)
		}
	}

	if !strings.Contains(output, "echo hello") && !strings.Contains(output, "sleep 0.1") {
		t.Errorf("Markdown output missing command data")
	}

	if !strings.Contains(output, "ran **") {
		t.Errorf("Markdown output missing comparison data")
	}
}
