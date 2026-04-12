package output

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/miklosn/cmdperf/internal/benchmark"
)

type JSONWriter struct{}

type jsonStat struct {
	Command        string  `json:"command"`
	TotalRuns      int     `json:"total_runs"`
	SuccessfulRuns int     `json:"successful_runs"`
	ErrorCount     int     `json:"error_count"`
	NonZeroExits   int     `json:"non_zero_exits"`
	MinNs          int64   `json:"min_ns"`
	MaxNs          int64   `json:"max_ns"`
	MeanNs         int64   `json:"mean_ns"`
	MedianNs       int64   `json:"median_ns"`
	P50Ns          int64   `json:"p50_ns"`
	P95Ns          int64   `json:"p95_ns"`
	P99Ns          int64   `json:"p99_ns"`
	StdDevNs       int64   `json:"stddev_ns"`
	Throughput     float64 `json:"throughput_per_sec"`
	TargetRate     float64 `json:"target_rate"`
}

func (w *JSONWriter) Write(writer io.Writer, stats []*benchmark.CommandStats) error {
	out := make([]jsonStat, 0, len(stats))
	for _, s := range stats {
		nonZero := 0
		for code, count := range s.ExitCodes {
			if code != 0 {
				nonZero += count
			}
		}
		out = append(out, jsonStat{
			Command:        s.Command.Raw,
			TotalRuns:      s.TotalRuns,
			SuccessfulRuns: s.SuccessfulRuns,
			ErrorCount:     s.ErrorCount,
			NonZeroExits:   nonZero,
			MinNs:          s.Min.Nanoseconds(),
			MaxNs:          s.Max.Nanoseconds(),
			MeanNs:         s.Mean.Nanoseconds(),
			MedianNs:       s.Median.Nanoseconds(),
			P50Ns:          s.P50.Nanoseconds(),
			P95Ns:          s.P95.Nanoseconds(),
			P99Ns:          s.P99.Nanoseconds(),
			StdDevNs:       s.StdDev.Nanoseconds(),
			Throughput:     s.Throughput,
			TargetRate:     s.TargetRate,
		})
	}
	enc := json.NewEncoder(writer)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		return fmt.Errorf("failed to write JSON: %w", err)
	}
	return nil
}
