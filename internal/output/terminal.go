package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/fatih/color"
	"github.com/miklosn/cmdperf/internal/benchmark"
)

type TerminalWriter struct{}

func (w *TerminalWriter) Write(writer io.Writer, stats []*benchmark.CommandStats) error {
	headerColor := color.New(color.FgHiCyan, color.Bold).SprintFunc()
	subheaderColor := color.New(color.FgCyan).SprintFunc()
	commandColor := color.New(color.FgHiYellow, color.Bold).SprintFunc()
	labelColor := color.New(color.FgHiBlue).SprintFunc()
	valueColor := color.New(color.FgWhite).SprintFunc()
	comparisonColor := color.New(color.FgHiMagenta, color.Bold).SprintFunc()
	fasterColor := color.New(color.FgHiGreen, color.Bold).SprintFunc()
	slowerColor := color.New(color.FgHiRed, color.Bold).SprintFunc()
	errorColor := color.New(color.FgHiRed).SprintFunc()

	fmt.Fprintln(writer, "\n"+headerColor("✨ cmdperf - Command Performance Benchmarking ✨"))
	fmt.Fprintln(writer, strings.Repeat("━", 50))

	headerLine := fmt.Sprintf("\n  %-12s %-30s %-30s %-15s %-10s\n",
		"Runs",
		"Mean ± StdDev",
		"Range (min … max)",
		"Throughput",
		"Errors")

	fmt.Fprint(writer, subheaderColor(headerLine))
	fmt.Fprintln(writer, strings.Repeat("━", 100))

	for _, stat := range stats {
		fmt.Fprintf(writer, "\n%s %s\n",
			labelColor("Command:"),
			commandColor(stat.Command.Raw))

		throughput := "-"
		meanStdDev := "-"
		timeRange := "-"

		if stat.SuccessfulRuns > 0 {
			meanStr := FormatDuration(stat.Mean)
			stdDevStr := FormatDuration(stat.StdDev)
			minStr := FormatDuration(stat.Min)
			maxStr := FormatDuration(stat.Max)

			meanStdDev = fmt.Sprintf("%s ± %s", meanStr, stdDevStr)
			timeRange = fmt.Sprintf("%s … %s", minStr, maxStr)
			throughput = FormatThroughput(stat.Throughput)
		}

		errorStr := fmt.Sprintf("%d", stat.ErrorCount)
		if stat.ErrorCount > 0 {
			errorStr = errorColor(errorStr)
		}

		line := fmt.Sprintf("  %-12d %-30s %-30s %-15s %-10s\n",
			stat.TotalRuns,
			meanStdDev,
			timeRange,
			throughput,
			errorStr)

		fmt.Fprint(writer, valueColor(line))

		hasNonZeroExitCodes := false
		for exitCode, count := range stat.ExitCodes {
			if exitCode != 0 && count > 0 {
				hasNonZeroExitCodes = true
				break
			}
		}

		if hasNonZeroExitCodes {
			fmt.Fprintf(writer, "\n  %s\n", labelColor("Exit Codes:"))
			for exitCode, count := range stat.ExitCodes {
				if exitCode != 0 {
					fmt.Fprintf(writer, "    %s: %d\n",
						slowerColor(fmt.Sprintf("Exit %d", exitCode)),
						count)
				}
			}
		}
	}

	if len(stats) > 1 {
		fmt.Fprintln(writer, "\n"+comparisonColor("⚡ Comparison:"))

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
			fmt.Fprintf(writer, "  '%s'\n  %s %s\n  '%s'\n\n",
				commandColor(stat.Command.Raw),
				slowerColor(fmt.Sprintf("ran %.2fx slower than", ratio)),
				fasterColor("↓"),
				commandColor(fastestCmd))
		}
	}

	return nil
}
