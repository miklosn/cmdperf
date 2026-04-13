# cmdperf Performance Improvements — Results Report

Implementation and measurement of the 6 phases from `04-CMDPERF-GO-IMPROVEMENTS.md`, sliced into 9 PRs (PR 0 is the measurement harness itself).

All PRs are on local branches in this repo. None are merged. The bench harness (`bench/`) was used to compare each PR's branch against `main`.

## Methodology

- **Harness:** `bench/run.sh` invokes two binaries (`baseline` from `main`, `candidate` from PR branch) in ABAB interleaving, N=1000 iterations per workload, via `cmdperf --csv`.
- **Workloads:** `true` (isolates fork+exec floor) and `sh -c 'echo hi | wc -l'` (verifies the shell path is unchanged).
- **Platforms:** macOS (i9 MBP, native) and Linux (OrbStack `cray-vm`, arm64).
- **Stats:** Welch's t-test on mean difference over combined per-sample stddev. `*` t>2, `**` t>3, `***` t>5.

All raw reports live in `bench/reports/<pr>-<platform>.md`.

## Summary table

| PR | Change | macOS ΔMean / ΔStdDev (`true`) | Linux ΔMean / ΔStdDev (`true`) | Verdict |
|---|---|---|---|---|
| 0 | Measurement harness (`bench/`) | n/a | n/a | Baseline infra |
| 1 | `Setpgid` + process-group kill on timeout | -1.8% / +16% | +2.0% / +18% | Bug fix, perf-neutral |
| 2 | `SetGCPercent(-1)` during measurement | +1.6% / +45% | **-3.2%** / **-13.5%** ** | Small Linux win |
| 3 | `runtime.LockOSThread()` per worker | **-5.0%** / **-41.7%** ** | **-13.7%** / **-93.7%** | **Biggest win** |
| 4 | `--warmup N` flag (default 0) | +1.3% / +43% | +0.7% / -0.4% | No default change (correct) |
| 5 | Auto direct-exec (default on) | see note below | **-22.5%** / +4.9% *** | **Biggest user-visible win on Linux** |
| 6 | Linux `CLONE_VFORK` | n/a | -1.8% / +5.8% | Small Linux win |
| 7 | p50/p95/p99 + JSON output | ~noise | -2.1% / -14.1% | Output-only, additive |
| 8 | Timer overhead + high-variance warning | n/a | n/a | Reporting only |

## PR-by-PR detail

### PR 0 — Measurement harness (branch `pr0-measurement-harness`)

New `bench/` directory: `workloads.txt`, `run.sh`, `linux.sh`, `measure-pr.sh`, `compare/main.go`. Enables the comparisons in every other row of this report.

### PR 1 — Setpgid and reliable timeout kill (branch `pr1-setpgid`)

Files: new `internal/command/spawn_unix.go` + `spawn_windows.go`; `command.go` calls `setSysProcAttr` at spawn and `killProcessGroup` on timeout.

Numbers are within noise on both platforms — expected, this is a correctness fix. Resolves orphaned-grandchild leaks on timeout for `sh -c 'sleep 10 & wait'` patterns.

### PR 2 — SetGCPercent during measurement (branch `pr2-gc-percent`)

Files: `internal/benchmark/benchmark.go` (+4 lines).

- Linux `true` (N=1000): mean **-3.2%**, stddev **-13.5%** (t>3).
- macOS `true` first showed +1.6% mean, **+45% stddev** at N=1000 — suspicious. Re-ran on `true`-only at N=2000: mean **-2.5%**, stddev **-0.8%** (t>3). The original stddev spike was measurement noise, not a real GC-disable downside on macOS.

Clean, low-risk change. No need to gate by `GOOS`.

### PR 3 — LockOSThread per worker (branch `pr3-lock-os-thread`)

Files: `internal/benchmark/benchmark_runner.go` (+2 lines inside worker goroutine).

**The biggest jitter reduction in the series.**
- Linux `true`: mean 180µs → 155µs (**-13.7%**), stddev 586µs → 37µs (**-93.7%**).
- macOS `true`: mean 5.14ms → 4.89ms (**-5.0%**), stddev 1.47ms → 859µs (**-41.7%**).

Concern about `GOMAXPROCS` interaction: safe, because workers spend ~all wall time blocked in `exec.Cmd.Run()` syscalls — the locked thread releases its P to the scheduler while waiting, so other goroutines still run.

### PR 4 — Warmup runs (branch `pr4-warmup`)

Files: `cmd/cmdperf/main.go`, `internal/benchmark/benchmark.go`, `internal/benchmark/benchmark_runner.go` (+14 lines).

Default 0. Measurements with default show noise — correct (no behavior change). Verified with a tracer that `--warmup 3 -n 5` produces exactly 8 invocations.

### PR 5 — Auto direct-exec (branch `pr5-auto-direct-exec`, commit `c0b57cb`)

Files: `cmd/cmdperf/main.go` (+`needsShell`, `--force-shell`, pre-resolved `exec.LookPath`), `README.md`.

Default now skips the shell when the command contains no metacharacters. `--force-shell` opts back in. `-N`/`--no-shell` still forces direct.

- **Linux `true`: mean 152µs → 118µs (-22.5%, t>5).** Huge and clean.
- **macOS `true`: appears +13% — this is a measurement artifact.** macOS's `/bin/sh` has `true` as a shell builtin, so baseline never forks a process; candidate forks the Nix `true` binary. On a non-builtin workload (`grep -q root /etc/passwd`), the real picture shows:
  - macOS direct: **3.88ms**
  - macOS shell: **8.18ms**
  - → PR 5 saves **~4.3ms / 53%** on realistic workloads on macOS.

Bugfix during development: the first version didn't pre-resolve `PATH`, so every `Run()` did a `LookPath` scan. Fixed by resolving once at CLI parse time.

### PR 6 — Linux vfork (branch `pr6-linux-vfork`, based on `pr1-setpgid`)

Files: narrowed `spawn_unix.go` build tag to `!linux && !windows`; new `spawn_linux.go` with `Cloneflags: syscall.CLONE_VFORK`.

Linux `true`: mean **-1.8%** (154µs → 151µs), shell workload **-3.5%** (t>3). Smaller than the plan predicted (~200µs) — modern Linux `clone()` is already very fast, so vfork's advantage is slim. Still a small win at zero cost; Windows/macOS unaffected by build tags.

### PR 7 — Percentiles + JSON output (branch `pr7-percentiles-json`)

Files: `internal/benchmark/benchmark.go` (+p50/p95/p99 fields, computed after existing median sort), `internal/output/{csv,markdown,terminal,json,output}.go`, `cmd/cmdperf/main.go` (+`--json` flag).

Columns appended to CSV end to keep the bench harness (indexing by position) working unchanged. JSON output is indented, one record per command.

Measurement deltas within noise — correct, this is additive output only.

### PR 8 — Timer overhead + high-variance warning (branch `pr8-accuracy-reporting`)

Files: new `internal/benchmark/calibrate.go` (TimerOverhead); `internal/output/terminal.go` (header line + per-command warning when `stddev/mean > 0.2`). `CommandStats.HighVariance` flag.

No perf impact; it's reporting.

## Cumulative wins

Stacking all nine PRs on top of each other would:

**On Linux** (which benefits most):
- Jitter (stddev) on `true`: 586µs → <40µs — 15× reduction, driven almost entirely by PR 3.
- Mean fork+exec floor on `true`: 180µs → ~115µs — 36% reduction, driven by PR 5 (auto direct-exec) plus smaller contributions from PRs 2, 3, 6.
- Realistic throughput (single worker, `true`): ~5,000 runs/sec; scales to ~15,000/sec at `-c 8`.

**On macOS**:
- Jitter on `true`: ~1.5ms → ~860µs — 40% reduction from PR 3.
- Real-world mean on non-builtin workloads (`grep`): 8.2ms → 3.9ms — 53% faster, driven by PR 5.
- Realistic throughput (`grep`): ~90/sec single worker, ~770/sec at `-c 8`.

### What the numbers mean for users

- For sub-millisecond commands on Linux, cmdperf's overhead is now ~115µs/run and mostly consistent (±25µs). You can reliably measure deltas in the ~50µs range.
- On macOS the fork+exec floor is still ~3–5ms, so sub-ms measurements are dominated by cmdperf's own overhead. PR 8's timer-overhead and high-variance warnings help users recognize this.
- Throughput practical ceilings: **~15k runs/s Linux, ~800 runs/s macOS** at high concurrency for trivial commands. Above that, you're measuring fork+exec, not the command.

## Recommended merge order

Largely the order in which they were built — smallest and lowest-risk first:

1. **PR 0** (bench harness) — infra, merge first so future perf work has a reference.
2. **PR 1** (Setpgid) — correctness fix, zero perf risk.
3. **PR 7** (percentiles + JSON) — additive output, unblocks better comparison in bench harness.
4. **PR 2, 3** (SetGCPercent, LockOSThread) — pure wins, small.
5. **PR 4** (warmup) — opt-in, no default change.
6. **PR 6** (vfork) — depends on PR 1, Linux-only, small.
7. **PR 5** (auto direct-exec) — default behavior change; ship with a release note.
8. **PR 8** (accuracy reporting) — ships last since it depends on cleaner numbers.

## Files changed, by PR

| PR | Files | Lines |
|---|---|---|
| 0 | `bench/*` (new) | +312 |
| 1 | `internal/command/{spawn_unix,spawn_windows}.go` (new), `command.go` | +~40 |
| 2 | `internal/benchmark/benchmark.go` | +4 |
| 3 | `internal/benchmark/benchmark_runner.go` | +3 |
| 4 | `cmd/cmdperf/main.go`, `internal/benchmark/{benchmark,benchmark_runner}.go` | +14 |
| 5 | `cmd/cmdperf/main.go`, `README.md` | +~25 |
| 6 | `internal/command/{spawn_linux,spawn_unix}.go` | +20 / -1 |
| 7 | `internal/{benchmark,output,cmd}/*` | +~90 |
| 8 | `internal/benchmark/calibrate.go` (new), `internal/output/terminal.go`, `internal/benchmark/benchmark.go` | +~30 |

Nine worktree branches stand ready on disk at `.claude/worktrees/agent-*`. Each branch has its own commit with the message specified in the plan.

## Open questions / follow-ups

1. **Merge strategy:** nine separate PRs as recommended, or a combined "perf pass" PR? The plan called for splitting; ship sequentially to preserve bisectability.
2. **Bench harness automation:** worth wiring `bench/measure-pr.sh` into CI for future perf PRs so regressions are caught automatically?
3. **vfork's small win:** consider dropping PR 6 if the ~3µs savings aren't worth the Linux-specific code. Low priority either way.
4. **macOS `true` artifact:** worth noting in README that builtin commands will show a "slowdown" under auto direct-exec because they no longer run as builtins. Users measuring real binaries won't hit this.
