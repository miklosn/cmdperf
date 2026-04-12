// compare reads baseline and candidate result directories produced by
// bench/run.sh and prints a Markdown comparison table to stdout.
//
// Usage: go run ./bench/compare <baseline-dir> <candidate-dir>
package main

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type row struct {
	Command    string
	TotalRuns  int64
	Min        float64
	Max        float64
	Mean       float64
	Median     float64
	StdDev     float64
	Throughput float64
}

func readCSV(path string) (*row, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	recs, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return nil, err
	}
	if len(recs) < 2 {
		return nil, fmt.Errorf("no data rows in %s", path)
	}
	r := recs[1]
	if len(r) < 11 {
		return nil, fmt.Errorf("unexpected csv format in %s: %d cols", path, len(r))
	}
	parseI := func(s string) int64 { v, _ := strconv.ParseInt(s, 10, 64); return v }
	parseF := func(s string) float64 { v, _ := strconv.ParseFloat(s, 64); return v }
	return &row{
		Command:    r[0],
		TotalRuns:  parseI(r[1]),
		Min:        parseF(r[5]),
		Max:        parseF(r[6]),
		Mean:       parseF(r[7]),
		Median:     parseF(r[8]),
		StdDev:     parseF(r[9]),
		Throughput: parseF(r[10]),
	}, nil
}

func formatNs(ns float64) string {
	switch {
	case ns >= 1e9:
		return fmt.Sprintf("%.2fs", ns/1e9)
	case ns >= 1e6:
		return fmt.Sprintf("%.2fms", ns/1e6)
	case ns >= 1e3:
		return fmt.Sprintf("%.1fµs", ns/1e3)
	default:
		return fmt.Sprintf("%.0fns", ns)
	}
}

func deltaPct(before, after float64) string {
	if before == 0 {
		return "n/a"
	}
	d := (after - before) / before * 100
	sign := ""
	if d > 0 {
		sign = "+"
	}
	return fmt.Sprintf("%s%.1f%%", sign, d)
}

// Welch's t-statistic for independent samples.
// We only have summary stats per workload — we compare stddev/mean ratios
// and flag a significant delta when the mean difference exceeds 2 combined stddevs.
func significance(b, c *row) string {
	if b.TotalRuns == 0 || c.TotalRuns == 0 {
		return ""
	}
	combined := math.Sqrt((b.StdDev*b.StdDev)/float64(b.TotalRuns) +
		(c.StdDev*c.StdDev)/float64(c.TotalRuns))
	if combined == 0 {
		return ""
	}
	t := math.Abs(c.Mean-b.Mean) / combined
	switch {
	case t > 5:
		return "***"
	case t > 3:
		return "**"
	case t > 2:
		return "*"
	default:
		return ""
	}
}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "usage: %s <baseline-dir> <candidate-dir>\n", os.Args[0])
		os.Exit(2)
	}
	baselineDir, candidateDir := os.Args[1], os.Args[2]

	baselines := map[string]*row{}
	files, err := filepath.Glob(filepath.Join(baselineDir, "*.csv"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for _, f := range files {
		r, err := readCSV(f)
		if err != nil {
			fmt.Fprintln(os.Stderr, "warn:", err)
			continue
		}
		baselines[filepath.Base(f)] = r
	}

	candidates := map[string]*row{}
	files, _ = filepath.Glob(filepath.Join(candidateDir, "*.csv"))
	for _, f := range files {
		r, err := readCSV(f)
		if err != nil {
			fmt.Fprintln(os.Stderr, "warn:", err)
			continue
		}
		candidates[filepath.Base(f)] = r
	}

	names := make([]string, 0, len(baselines))
	for k := range baselines {
		names = append(names, k)
	}
	sort.Strings(names)

	fmt.Println("# cmdperf bench comparison")
	fmt.Println()
	fmt.Printf("Baseline dir: `%s`  \nCandidate dir: `%s`\n\n", baselineDir, candidateDir)
	fmt.Println("Significance: `*` t>2, `**` t>3, `***` t>5 (Welch's t on per-sample error of the mean).")
	fmt.Println()
	fmt.Println("| Workload | Mean (b → c) | ΔMean | StdDev (b → c) | ΔStdDev | Sig |")
	fmt.Println("|---|---|---|---|---|---|")
	for _, n := range names {
		b := baselines[n]
		c, ok := candidates[n]
		if !ok {
			continue
		}
		cmd := b.Command
		if len(cmd) > 30 {
			cmd = cmd[:27] + "..."
		}
		cmd = strings.ReplaceAll(cmd, "|", "\\|")
		fmt.Printf("| `%s` | %s → %s | %s | %s → %s | %s | %s |\n",
			cmd,
			formatNs(b.Mean), formatNs(c.Mean), deltaPct(b.Mean, c.Mean),
			formatNs(b.StdDev), formatNs(c.StdDev), deltaPct(b.StdDev, c.StdDev),
			significance(b, c),
		)
	}
}
