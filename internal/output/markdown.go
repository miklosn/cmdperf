package output

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/miklosn/cmdperf/internal/benchmark"
)

type MarkdownWriter struct{}

func (w *MarkdownWriter) Write(writer io.Writer, stats []*benchmark.CommandStats) error {
	bufWriter := bufio.NewWriter(writer)
	defer bufWriter.Flush()

	fmt.Fprintf(bufWriter, "# ✨ cmdperf - Command Performance Benchmarking ✨\n\n")
	fmt.Fprintf(bufWriter, "Generated on: %s\n\n", time.Now().Format(time.RFC1123))

	fmt.Fprintf(bufWriter, "## Benchmark Parameters\n\n")

	if len(stats) > 0 {
		cmd := stats[0].Command
		fmt.Fprintf(bufWriter, "- **Total Commands**: %d\n", len(stats))
		fmt.Fprintf(bufWriter, "- **Runs per Command**: %d\n", stats[0].TotalRuns)
		fmt.Fprintf(bufWriter, "- **Parallelism**: %d\n", cmd.Parallelism)
		fmt.Fprintf(bufWriter, "- **Timeout**: %s\n", cmd.Timeout)
		fmt.Fprintf(bufWriter, "- **Shell**: %s\n", cmd.Shell)
		fmt.Fprintf(bufWriter, "- **Shell Options**: %s\n", strings.Join(cmd.ShellOptions, " "))
		if stats[0].TargetRate > 0 {
			fmt.Fprintf(bufWriter, "- **Target Rate**: %.2f req/sec\n", stats[0].TargetRate)
		}
		fmt.Fprintln(bufWriter)
	}

	fmt.Fprintf(bufWriter, "## Summary\n\n")
	fmt.Fprintf(bufWriter, "| Command | Runs | Mean ± StdDev | Min | Max | Throughput | Rate | Errors |\n")
	fmt.Fprintf(bufWriter, "|---------|------|--------------|-----|-----|------------|------|-------|\n")

	for _, stat := range stats {
		meanStr := FormatDuration(stat.Mean)
		stdDevStr := FormatDuration(stat.StdDev)
		minStr := FormatDuration(stat.Min)
		maxStr := FormatDuration(stat.Max)
		throughputStr := FormatThroughput(stat.Throughput)

		escapedCmd := strings.ReplaceAll(stat.Command.Raw, "|", "\\|")

		rateStr := "-"
		if stat.TargetRate > 0 {
			rateStr = fmt.Sprintf("%.2f/%.2f", stat.Throughput, stat.TargetRate)
		}

		fmt.Fprintf(bufWriter, "| `%s` | %d | %s ± %s | %s | %s | %s | %s | %d |\n",
			escapedCmd,
			stat.TotalRuns,
			meanStr,
			stdDevStr,
			minStr,
			maxStr,
			throughputStr,
			rateStr,
			stat.ErrorCount)
	}

	fmt.Fprintf(bufWriter, "\n## Command Parameters\n\n")

	for i, stat := range stats {
		escapedCmd := strings.ReplaceAll(stat.Command.Raw, "`", "\\`")

		fmt.Fprintf(bufWriter, "### Command %d: `%s`\n\n", i+1, escapedCmd)

		fmt.Fprintf(bufWriter, "- **Parallelism**: %d\n", stat.Command.Parallelism)
		fmt.Fprintf(bufWriter, "- **Timeout**: %s\n", stat.Command.Timeout)
		fmt.Fprintf(bufWriter, "- **Shell**: %s\n", stat.Command.Shell)
		fmt.Fprintf(bufWriter, "- **Shell Options**: %s\n", strings.Join(stat.Command.ShellOptions, " "))

		if stat.ErrorCount > 0 {
			fmt.Fprintf(bufWriter, "- **Error Count**: %d\n", stat.ErrorCount)
		}

		if len(stat.ExitCodes) > 0 {
			fmt.Fprintf(bufWriter, "- **Exit Codes**:\n")
			for exitCode, count := range stat.ExitCodes {
				if exitCode == 0 {
					fmt.Fprintf(bufWriter, "  - Success (0): %d\n", count)
				} else {
					fmt.Fprintf(bufWriter, "  - Error (%d): %d\n", exitCode, count)
				}
			}
		}
	}

	if len(stats) > 1 {
		fmt.Fprintf(bufWriter, "\n## Comparison\n\n")

		fastestIdx := 0
		for i := 1; i < len(stats); i++ {
			if stats[i].Mean < stats[fastestIdx].Mean {
				fastestIdx = i
			}
		}

		fastestCmd := stats[fastestIdx].Command.Raw
		escapedFastestCmd := strings.ReplaceAll(fastestCmd, "`", "\\`")

		for i, stat := range stats {
			if i == fastestIdx {
				continue
			}

			escapedCmd := strings.ReplaceAll(stat.Command.Raw, "`", "\\`")
			ratio := float64(stat.Mean) / float64(stats[fastestIdx].Mean)

			fmt.Fprintf(bufWriter, "- `%s` ran **%.2fx slower** than `%s`\n",
				escapedCmd, ratio, escapedFastestCmd)
		}
	}

	return nil
}
