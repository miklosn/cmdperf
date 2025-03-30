package output

import (
	"encoding/csv"
	"fmt"
	"io"

	"github.com/miklosn/cmdperf/internal/benchmark"
)

type CSVWriter struct{}

func (w *CSVWriter) Write(writer io.Writer, stats []*benchmark.CommandStats) error {
	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	header := []string{
		"Command",
		"TotalRuns",
		"SuccessfulRuns",
		"ErrorCount",
		"NonZeroExitCodes",
		"Min (ns)",
		"Max (ns)",
		"Mean (ns)",
		"Median (ns)",
		"StdDev (ns)",
		"Throughput (/s)",
	}
	if err := csvWriter.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	for _, stat := range stats {
		nonZeroExitCodes := 0
		for exitCode, count := range stat.ExitCodes {
			if exitCode != 0 {
				nonZeroExitCodes += count
			}
		}

		row := []string{
			stat.Command.Raw,
			fmt.Sprintf("%d", stat.TotalRuns),
			fmt.Sprintf("%d", stat.SuccessfulRuns),
			fmt.Sprintf("%d", stat.ErrorCount),
			fmt.Sprintf("%d", nonZeroExitCodes),
			fmt.Sprintf("%d", stat.Min.Nanoseconds()),
			fmt.Sprintf("%d", stat.Max.Nanoseconds()),
			fmt.Sprintf("%d", stat.Mean.Nanoseconds()),
			fmt.Sprintf("%d", stat.Median.Nanoseconds()),
			fmt.Sprintf("%d", stat.StdDev.Nanoseconds()),
			fmt.Sprintf("%f", stat.Throughput),
		}
		if err := csvWriter.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row for command '%s': %w", stat.Command.Raw, err)
		}
	}

	csvWriter.Flush()
	if err := csvWriter.Error(); err != nil {
		return fmt.Errorf("error flushing CSV data: %w", err)
	}

	return nil
}
