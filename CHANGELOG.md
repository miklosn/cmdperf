# Changelog

All notable changes to cmdperf are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

Targeted at 0.2.0. Adds structured output, percentile reporting and
measurement-accuracy context, makes cmdperf work on Windows, and fixes several
cases where reported statistics were wrong.

### Added

- `--json=<file>` writes results as structured JSON, for CI pipelines and
  programmatic consumption ([#14](https://github.com/miklosn/cmdperf/pull/14)).
- p50/p95/p99 percentiles in terminal, CSV, Markdown and JSON output. The new
  columns are appended at the end of each CSV/Markdown row, so existing
  consumers that index by column position are unaffected
  ([#14](https://github.com/miklosn/cmdperf/pull/14)).
- The UI header shows the measured timer overhead (the cost of one
  `time.Since` call) as context for sub-microsecond deltas
  ([#20](https://github.com/miklosn/cmdperf/pull/20)).
- Results with high variance (stddev above 20% of mean) are flagged with a
  warning suggesting more runs ([#20](https://github.com/miklosn/cmdperf/pull/20)).

### Fixed

- cmdperf was unusable on Windows: the default shell was hardcoded to
  `/bin/sh`, so every spawn failed. The default now resolves at runtime to
  `%COMSPEC%` (cmd.exe) with `/c` on Windows and `/bin/sh -c` elsewhere;
  explicit `--shell`/`--shell-opt` behavior is unchanged
  ([#19](https://github.com/miklosn/cmdperf/pull/19)).
- Commands that never started (`cmd.Start` failure) were counted as successful
  runs and fed near-zero durations into min/mean/stddev, producing garbage
  statistics. They are now excluded from both the success count and the timing
  stats ([#18](https://github.com/miklosn/cmdperf/pull/18)).
- Commands that always exit non-zero (for example `/bin/false`) reported
  `stddev` of 0 while still reporting a real mean, because stddev skipped every
  result with an error. Stddev now skips only timed-out results, matching how
  mean is calculated.
- Grandchild processes survived a timeout: `sh -c 'sleep 10 & wait'` left the
  inner `sleep` running. Spawned commands now get their own process group,
  which is killed as a group when the timeout expires
  ([#14](https://github.com/miklosn/cmdperf/pull/14)).
- The three horizontal rules in the inline UI used different width caps, so the
  top and bottom rules were shorter than the middle one.

### Performance

Measurement-accuracy work, quantified by the new benchmark harness
([#14](https://github.com/miklosn/cmdperf/pull/14)):

- The garbage collector is disabled for the duration of the measurement window
  and restored afterwards, removing GC pauses of 50-500 microseconds from
  timing samples. Peak memory stays bounded by the existing 1000-slot sample
  buffer.
- Benchmark workers are pinned to OS threads on Linux and macOS, so samples are
  not perturbed by the Go scheduler migrating goroutines between threads.
  Pinning is deliberately disabled on FreeBSD, where it regressed mean by
  5-15% and stddev by 25-107%.

### Internal

- Dev tooling moved from devenv/nix to [mise](https://mise.jdx.dev), with exact
  tool versions pinned in `mise.toml` and a `.githooks/pre-commit` hook wired
  up by `mise run setup`.
- Added `.golangci.yml` so the newer golangci-lint installed by mise applies
  the same pass/fail criteria as the version previously pinned by nix.
- Added `bench/`, a harness that drives two cmdperf binaries through a fixed
  workload suite and emits a Markdown comparison report, plus a cross-platform
  benchmark workflow covering Linux, macOS, FreeBSD, Windows, OpenBSD and
  NetBSD.
- The benchmark harness accepts a `WORKLOADS_FILE` override and ships a
  cmd.exe-syntax workload set, used by the Windows benchmark job
  ([#19](https://github.com/miklosn/cmdperf/pull/19)).
- Added a workflow benchmarking cmdperf across Go toolchain versions
  (1.22-1.26 against the pinned 1.24.13) on ubuntu x86/arm and macos
  ([#21](https://github.com/miklosn/cmdperf/pull/21)).
- Added `bench/vm/freebsd.sh`, which benchmarks FreeBSD in a local QEMU+HVF
  VM. A golden base image is provisioned once and cached; each run boots a
  fresh copy-on-write overlay in about 30s. This puts the FreeBSD spawn floor
  at roughly 960us with 80us stddev, against 1.5-1.8ms with 0.6-1.4ms stddev
  on GitHub-hosted QEMU runners ([#22](https://github.com/miklosn/cmdperf/pull/22)).
- Added `bench/gcp/run.sh`, which runs the A/B/A/B harness on a dedicated GCP
  spot instance (c4d x86 or c4a arm) with a pinned Go toolchain and tears the
  instance down afterwards. Identical binaries agree within +/-0.8% on c4d,
  making it the cleanest measurement environment in the project
  ([#22](https://github.com/miklosn/cmdperf/pull/22)).

## [0.1.4] - 2026-01-21

### Fixed

- Exit with a non-zero status when writing an output file fails.

### Changed

- Optimized duration-based benchmarking.

## [0.1.3] - 2025-08-01

Release-tooling only; no changes to cmdperf itself.

## [0.1.2] - 2025-08-01

### Added

- Homebrew cask distribution via `miklosn/homebrew-tap`.

## [0.1.0] - 2025-04-07

### Added

- Per-thread rate limiting, with achieved rate reported in the output.

## [0.0.1] - 2025-04-02

Initial release.

[Unreleased]: https://github.com/miklosn/cmdperf/compare/v0.1.4...HEAD
[0.1.4]: https://github.com/miklosn/cmdperf/compare/v0.1.3...v0.1.4
[0.1.3]: https://github.com/miklosn/cmdperf/compare/v0.1.2...v0.1.3
[0.1.2]: https://github.com/miklosn/cmdperf/compare/v0.1.0...v0.1.2
[0.1.0]: https://github.com/miklosn/cmdperf/compare/v0.0.1...v0.1.0
[0.0.1]: https://github.com/miklosn/cmdperf/releases/tag/v0.0.1
