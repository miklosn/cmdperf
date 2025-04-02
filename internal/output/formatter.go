package output

import (
	"fmt"
	"strings"
	"time"

	"github.com/miklosn/cmdperf/internal/benchmark"
)

type TableColumn struct {
	Header string
	Width  int
	Value  func(stat *benchmark.CommandStats) string
}

type RenderOptions struct {
	ShowProgress    bool
	ShowComparison  bool
	ProgressPercent float64
	Elapsed         time.Duration
	ETA             time.Duration
	TotalRuns       int
	Duration        time.Duration
	IsCancelled     bool
	IsFinished      bool
	ColorizeFunc    func(string, string) string
}

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

// GetStandardColumns returns the standard columns for the benchmark output
func GetStandardColumns(stats []*benchmark.CommandStats, termWidth int, showIterations bool) []TableColumn {
	// Determine if rate limiting is active
	rateActive := len(stats) > 0 && stats[0] != nil && stats[0].TargetRate > 0

	// Calculate dynamic column widths based on terminal width
	runsWidth := max(min(int(float64(termWidth)*0.10), 15), 8)
	meanWidth := max(min(int(float64(termWidth)*0.25), 30), 15)
	rangeWidth := max(min(int(float64(termWidth)*0.25), 30), 15)
	throughputWidth := max(min(int(float64(termWidth)*0.15), 20), 10)
	errorsWidth := max(min(int(float64(termWidth)*0.15), 15), 8)

	// For very narrow terminals, adjust columns
	if termWidth < 80 {
		rangeWidth = 0
		if termWidth < 60 {
			throughputWidth = 0
			if termWidth < 40 {
				errorsWidth = 0
				meanWidth = max(meanWidth-5, 10)
			}
		}
	}

	columns := []TableColumn{}

	// Runs column
	columns = append(columns, TableColumn{
		Header: "Runs",
		Width:  runsWidth,
		Value: func(stat *benchmark.CommandStats) string {
			if stat == nil {
				return "-"
			}
			if showIterations {
				return fmt.Sprintf("%d/%d", stat.TotalRuns, stat.Command.Parallelism)
			}
			return fmt.Sprintf("%d", stat.TotalRuns)
		},
	})

	// Mean ± StdDev column
	if meanWidth > 0 {
		columns = append(columns, TableColumn{
			Header: "Mean ± StdDev",
			Width:  meanWidth,
			Value: func(stat *benchmark.CommandStats) string {
				if stat == nil || stat.SuccessfulRuns == 0 {
					return "-"
				}
				meanStr := FormatDuration(stat.Mean)
				stdDevStr := FormatDuration(stat.StdDev)
				return fmt.Sprintf("%s ± %s", meanStr, stdDevStr)
			},
		})
	}

	// Range column
	if rangeWidth > 0 {
		columns = append(columns, TableColumn{
			Header: "Range (min … max)",
			Width:  rangeWidth,
			Value: func(stat *benchmark.CommandStats) string {
				if stat == nil || stat.SuccessfulRuns == 0 {
					return "-"
				}
				minStr := FormatDuration(stat.Min)
				maxStr := FormatDuration(stat.Max)
				return fmt.Sprintf("%s … %s", minStr, maxStr)
			},
		})
	}

	// Throughput column
	if throughputWidth > 0 {
		header := "Throughput"
		if rateActive {
			header = "Throughput (target)"
		}

		columns = append(columns, TableColumn{
			Header: header,
			Width:  throughputWidth,
			Value: func(stat *benchmark.CommandStats) string {
				if stat == nil || stat.SuccessfulRuns == 0 {
					return "-"
				}
				if stat.TargetRate > 0 {
					return FormatThroughputWithRate(stat.Throughput, stat.TargetRate)
				}
				return FormatThroughput(stat.Throughput)
			},
		})
	}

	// Errors column
	if errorsWidth > 0 {
		columns = append(columns, TableColumn{
			Header: "Errors",
			Width:  errorsWidth,
			Value: func(stat *benchmark.CommandStats) string {
				if stat == nil || stat.ErrorCount == 0 {
					return "-"
				}

				errorStr := fmt.Sprintf("%d", stat.ErrorCount)

				// Add exit code information for non-zero exit codes
				timeoutInfo := ""
				for exitCode, count := range stat.ExitCodes {
					if exitCode != 0 {
						timeoutInfo += fmt.Sprintf(" [Exit %d: %d]", exitCode, count)
					}
				}

				if timeoutInfo != "" {
					return fmt.Sprintf("%s%s", errorStr, timeoutInfo)
				}

				return errorStr
			},
		})
	}

	return columns
}

// RenderTable renders a table of benchmark results
func RenderTable(stats []*benchmark.CommandStats, options RenderOptions) string {
	var output strings.Builder

	// Get terminal width (use a reasonable default if not available)
	termWidth := 100 // Default width

	// Get standard columns
	columns := GetStandardColumns(stats, termWidth, options.TotalRuns > 0)

	// Render header
	output.WriteString(options.ColorizeFunc("✨ cmdperf - Command Performance Benchmarking ✨\n", "header"))
	output.WriteString(strings.Repeat("─", min(termWidth-5, 60)) + "\n\n")

	// Render column headers
	headerFormat := ""
	headerArgs := []interface{}{}

	for _, col := range columns {
		headerFormat += fmt.Sprintf("  %%-%ds ", col.Width)
		headerArgs = append(headerArgs, col.Header)
	}
	headerFormat += "\n"

	output.WriteString(options.ColorizeFunc(fmt.Sprintf(headerFormat, headerArgs...), "subheader"))
	output.WriteString(strings.Repeat("━", min(termWidth-5, 100)) + "\n")

	// Render each command's stats
	for _, stat := range stats {
		if stat == nil || stat.Command == nil {
			continue
		}

		// Render command name
		cmdName := stat.Command.Raw
		if options.Duration > 0 {
			output.WriteString(fmt.Sprintf("%s %s %s\n",
				options.ColorizeFunc("Command:", "label"),
				options.ColorizeFunc(cmdName, "command"),
				options.ColorizeFunc(fmt.Sprintf("(running for %s)", options.Duration), "subheader")))
		} else {
			output.WriteString(fmt.Sprintf("%s %s\n",
				options.ColorizeFunc("Command:", "label"),
				options.ColorizeFunc(cmdName, "command")))
		}

		// Render command stats
		lineFormat := ""
		lineArgs := []interface{}{}

		for _, col := range columns {
			lineFormat += fmt.Sprintf("  %%-%ds ", col.Width)
			lineArgs = append(lineArgs, col.Value(stat))
		}
		lineFormat += "\n"

		output.WriteString(options.ColorizeFunc(fmt.Sprintf(lineFormat, lineArgs...), "value"))

		// Render exit codes if any
		if HasNonZeroExitCodes(stat) {
			output.WriteString(fmt.Sprintf("\n  %s\n", options.ColorizeFunc("Exit Codes:", "label")))
			for exitCode, count := range stat.ExitCodes {
				if exitCode != 0 {
					output.WriteString(fmt.Sprintf("    %s: %d\n",
						options.ColorizeFunc(fmt.Sprintf("Exit %d", exitCode), "error"),
						count))
				}
			}
		}
	}

	// Render progress information if requested
	if options.ShowProgress {
		output.WriteString(strings.Repeat("─", min(termWidth-5, 60)) + "\n")

		// Create a progress bar
		barWidth := min(30, termWidth/3)
		completedWidth := int(float64(barWidth) * options.ProgressPercent)

		var progressBar string
		if completedWidth < 0 {
			completedWidth = 0
		}
		remainingWidth := barWidth - completedWidth
		if remainingWidth < 0 {
			remainingWidth = 0
		}

		progressBar = fmt.Sprintf("[%s%s] %.0f%%",
			strings.Repeat("█", completedWidth),
			strings.Repeat("░", remainingWidth),
			options.ProgressPercent*100)

		// Format ETA
		etaStr := "calculating..."
		if options.ProgressPercent >= 1.0 {
			etaStr = "done"
		} else if options.ETA > 0 {
			etaStr = options.ETA.String()
		}

		output.WriteString(fmt.Sprintf("%s %s | %s %s | %s %s\n",
			options.ColorizeFunc("Elapsed:", "label"),
			options.ColorizeFunc(options.Elapsed.String(), "value"),
			options.ColorizeFunc("Progress:", "label"),
			options.ColorizeFunc(progressBar, "progress"),
			options.ColorizeFunc("ETA:", "label"),
			options.ColorizeFunc(etaStr, "value")))

		// Add status message
		if options.IsFinished {
			if options.IsCancelled {
				output.WriteString("\n" + options.ColorizeFunc("⚠️  Benchmark cancelled!", "error") + "\n")
			} else {
				output.WriteString("\n" + options.ColorizeFunc("✅ Benchmark completed!", "completed") + "\n")
			}
		} else {
			output.WriteString("\n" + options.ColorizeFunc("Press Ctrl+C to interrupt", "value") + "\n")
		}
	}

	// Render comparison if multiple commands and requested
	if options.ShowComparison && len(stats) > 1 {
		output.WriteString("\n" + options.ColorizeFunc("⚡ Comparison:", "comparison") + "\n")

		// Find fastest command
		fastestIdx := 0
		for i := 1; i < len(stats); i++ {
			if stats[i].Mean < stats[fastestIdx].Mean {
				fastestIdx = i
			}
		}

		fastestCmd := stats[fastestIdx].Command.Raw

		for i, stat := range stats {
			if i == fastestIdx {
				continue
			}

			ratio := float64(stat.Mean) / float64(stats[fastestIdx].Mean)
			output.WriteString(fmt.Sprintf("  '%s'\n  %s %s\n  '%s'\n\n",
				options.ColorizeFunc(stat.Command.Raw, "command"),
				options.ColorizeFunc(fmt.Sprintf("ran %.2fx slower than", ratio), "slower"),
				options.ColorizeFunc("↓", "faster"),
				options.ColorizeFunc(fastestCmd, "command")))
		}
	}

	return output.String()
}

// Helper functions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
