package ui

import (
	"os"
	"time"

	"github.com/miklosn/cmdperf/internal/output"
	"golang.org/x/term"
)

func formatDuration(d time.Duration) string {
	return output.FormatDuration(d)
}

func formatThroughputWithRate(throughput, targetRate float64) string {
	return output.FormatThroughputWithRate(throughput, targetRate)
}

func formatThroughput(throughput float64) string {
	return output.FormatThroughput(throughput)
}

func getTerminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width <= 0 {
		return 80
	}
	return width
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
