# cmdperf Performance Harness

Measures and compares two cmdperf binaries (`baseline` vs `candidate`) across a fixed workload suite. Runs on macOS (host) and Linux (OrbStack).

## Usage

```
# Build both binaries
make -C .. build           # baseline: use current main
cp ../cmdperf bench/bin/baseline
git checkout <pr-branch>
make -C .. build
cp ../cmdperf bench/bin/candidate

# Run on macOS
bench/run.sh bench/bin/baseline bench/bin/candidate macos

# Run on OrbStack Linux
bench/linux.sh bench/bin/baseline bench/bin/candidate linux

# Compare (writes Markdown report)
go run ./bench/compare bench/results/<run-id>/baseline bench/results/<run-id>/candidate > report.md
```

## Layout

- `workloads.txt` — one command per line; `#` starts a comment
- `run.sh` — macOS/generic driver
- `linux.sh` — wraps `run.sh` for OrbStack `cray-vm`
- `compare/` — Go program that reads baseline + candidate CSVs and produces a Markdown diff
- `results/` — output (gitignored)
- `bin/` — built binaries (gitignored)

## Methodology

- N=2000 iterations per workload
- ABAB interleaving (run baseline, candidate, baseline, candidate) to smooth out thermal drift
- Single-CPU pin via `taskset -c 0` on Linux where available
- Reports: mean, stddev, p50, p95, p99, throughput, and delta vs baseline with Welch's t-test
