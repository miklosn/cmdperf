package ui

import (
	"fmt"
	"os"
	"time"

	"golang.org/x/term"
)

func formatDuration(d time.Duration) string {
	if d < time.Microsecond {
		return fmt.Sprintf("%.2f ns", float64(d.Nanoseconds()))
	} else if d < time.Millisecond {
		return fmt.Sprintf("%.2f µs", float64(d.Nanoseconds())/1000)
	} else if d < time.Second {
		return fmt.Sprintf("%.2f ms", float64(d.Nanoseconds())/1000000)
	} else {
		return fmt.Sprintf("%.2f s", d.Seconds())
	}
}

func formatThroughput(throughput float64) string {
	if throughput <= 0 {
		return "-"
	}

	// Format based on magnitude
	if throughput >= 1000000 {
		return fmt.Sprintf("%.2f M/s", throughput/1000000)
	} else if throughput >= 1000 {
		return fmt.Sprintf("%.2f K/s", throughput/1000)
	} else if throughput >= 1 {
		return fmt.Sprintf("%.2f /s", throughput)
	} else if throughput >= 1.0/60 {
		return fmt.Sprintf("%.2f /min", throughput*60)
	} else {
		return fmt.Sprintf("%.2f /hr", throughput*3600)
	}
}

func getTerminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width <= 0 {
		return 80 // Default fallback width
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
