package output

import (
	"fmt"
	"time"

	"github.com/miklosn/cmdperf/internal/benchmark"
)

// FormatDuration formats a duration in a human-readable form with appropriate units
func FormatDuration(d time.Duration) string {
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

// FormatThroughput formats throughput in a human-readable form with appropriate units
func FormatThroughput(throughput float64) string {
	if throughput <= 0 {
		return "-"
	}
	if throughput >= 1000000 {
		return fmt.Sprintf("%.2f M/s", throughput/1000000)
	} else if throughput >= 1000 {
		return fmt.Sprintf("%.2f K/s", throughput/1000)
	} else if throughput >= 0.5 {
		return fmt.Sprintf("%.2f /s", throughput)
	} else if throughput >= 0.01 {
		return fmt.Sprintf("%.2f /min", throughput*60)
	} else {
		return fmt.Sprintf("%.2f /hr", throughput*3600)
	}
}

// FormatThroughputWithRate formats throughput and includes rate information if available
func FormatThroughputWithRate(throughput, targetRate float64) string {
	// Format the base throughput
	baseStr := FormatThroughput(throughput)

	// If target rate is set, include only the target rate
	if targetRate > 0 {
		return fmt.Sprintf("%s (%.0f)", baseStr, targetRate)
	}

	return baseStr
}

// GetThroughputHeader returns the appropriate header text for throughput column
func GetThroughputHeader(stats []*benchmark.CommandStats) string {
	if len(stats) > 0 && stats[0] != nil && stats[0].TargetRate > 0 {
		return "Throughput (target)"
	}
	return "Throughput"
}

// HasNonZeroExitCodes checks if a command has any non-zero exit codes
func HasNonZeroExitCodes(stat *benchmark.CommandStats) bool {
	if stat == nil {
		return false
	}

	for exitCode, count := range stat.ExitCodes {
		if exitCode != 0 && count > 0 {
			return true
		}
	}

	return false
}
