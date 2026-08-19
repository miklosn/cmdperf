# Changelog

All notable changes to cmdperf are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.0]

This release focuses on performance, thanks to several optimizations, especially
on Linux, cmdperf is now indeed as good for high frequency benchmarking as anything.

### Added

- `--json=<file>` writes results as structured JSON, for CI pipelines and
  programmatic consumption.
- The UI header shows the measured timer overhead.
- Results with high variance (stddev above 20% of mean) are flagged with a
  warning suggesting more runs.

### Fixed

- cmdperf was unusable on Windows: the default shell was hardcoded to
  `/bin/sh`, so every spawn failed. The default now resolves at runtime to
  `%COMSPEC%` (cmd.exe) with `/c` on Windows and `/bin/sh -c` elsewhere;
  explicit `--shell`/`--shell-opt` behavior is unchanged.
- Commands that never started (`cmd.Start` failure) were counted as successful
  runs and fed near-zero durations into min/mean/stddev, producing garbage
  statistics.
- Commands that always exit non-zero (for example `/bin/false`) reported
  `stddev` of 0 while still reporting a real mean, because stddev skipped every
  result with an error. Stddev now skips only timed-out results, matching how
  mean is calculated.
- Grandchild processes survived a timeout: `sh -c 'sleep 10 & wait'` left the
  inner `sleep` running. Spawned commands now get their own process group,
  which is killed as a group when the timeout expires.
- The three horizontal rules in the inline UI used different width caps, so the
  top and bottom rules were shorter than the middle one.

### Internal

- Dev tooling moved from devenv/nix to [mise](https://mise.jdx.dev).
- Added `.golangci.yml` so the newer golangci-lint installed by mise applies
  the same pass/fail criteria as the version previously pinned by nix.
- Added `bench/`, a harness that drives two cmdperf binaries through a fixed
  workload suite covering Linux, macOS, FreeBSD, Windows, OpenBSD and
  NetBSD.

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
