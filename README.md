# ✨ cmdperf - Command Performance Benchmarking ✨

A command-line benchmarking tool that allows you to measure and compare the performance of shell commands.

![cmdperf demo](doc/demo.gif)

## Features

- Benchmark shell commands
- Support for parallel execution
- Compare commands
- Real-time terminal UI with live statistics, progress tracking and ETA
- Responsive
- Color schemes :)

## Non-Goals

- Precise analysis. We're deliberately spawning shells and such. cmdperf is a tool for quick and dirty benchmarking, not a scientific tool.
- Not a replacement for specialized benchmarking tools.

## Installation

```bash
go install github.com/miklosn/cmdperf/cmd/cmdperf@latest
```

## Usage

```bash
cmdperf [options] <command...>
```

For example:

```bash
# Basic usage
cmdperf "sleep 0.1"

# Multiple commands to compare
cmdperf "sleep 0.1" "sleep 0.2"

# Parallel execution with 10 concurrent processes
cmdperf -c 10 "curl -s https://example.com > /dev/null"

# Run 100 iterations of each command
cmdperf -n 100 "redis-cli PING"

# Run benchmark for 30 seconds
cmdperf -d 30s "redis-cli PING"
```

## CLI Options

```
Arguments:
  <command...>    Command(s) to benchmark

Options:
  -n, --runs=<n>                Number of runs to perform [default: 10]
  -c, --concurrency=<n>         Number of concurrent executions [default: 1]
      --color-scheme=<scheme>   Color scheme to use (auto, catppuccin, tokyonight, nord, monokai, solarized, solarized-light, gruvbox, monochrome) [default: auto]
      --list-color-schemes      List available color schemes
  -t, --timeout=<duration>      Timeout for each command execution [default: 1m]
  -d, --duration=<duration>     Total benchmark duration (overrides --runs)
  -s, --shell=<shell>           Shell to use for command execution [default: /bin/sh]
      --shell-opt=<opt>         Shell option (can be repeated) [default: -c]
  -N, --no-shell                Execute commands directly without a shell
      --csv=<file>              Write results to CSV file
      --markdown=<file>         Write results to Markdown file
      --version                 Show version information
      --fail-on-error           Exit with non-zero status if any command returns non-zero exit code
      --cpu-profile=<file>      Write CPU profile to file
      --mem-profile=<file>      Write memory profile to file
      --block-profile=<file>    Write goroutine blocking profile to file
      --pprof-server            Start pprof HTTP server on :6060
```

## Color Schemes

cmdperf supports various color schemes to match your terminal theme:

```bash
# Use a specific color scheme
cmdperf --color-scheme=nord "sleep 0.1"

# Automatically detect terminal background and choose appropriate theme
cmdperf --color-scheme=auto "sleep 0.1"

# List available color schemes
cmdperf --list-color-schemes
```

Available color schemes include:

- default: Default color scheme
- auto: Automatically selects a theme based on terminal background
- catppuccin: Soothing pastel theme (Mocha variant)
- tokyonight: A dark and elegant theme
- nord: Arctic, north-bluish color palette
- monokai: Vibrant and colorful theme
- solarized: Precision colors for machines and people (dark variant)
- solarized-light: Precision colors for machines and people (light variant)
- monochrome: Simple black and white theme (no colors)

## Direct Execution Mode

By default, cmdperf executes commands through a shell (usually `/bin/sh -c`). This allows for shell features like pipes, redirections, and variable expansions. However, for simple commands, you can use direct execution mode to bypass the shell:

```bash
cmdperf -N "ls -la"
```

In direct execution mode:

- The command is split by spaces (respecting quotes)
- The first part is used as the executable
- The remaining parts are passed as arguments
- Shell features like pipes (`|`), redirections (`>`), and variable expansions (`$VAR`) won't work
- The command is executed directly without a shell

Please note that even without spawning a shell, `cmdperf` is not designed for high frequency benchmarking.

## Output

cmdperf provides a colorful, real-time UI that shows:

- Command execution progress
- Mean execution time with standard deviation
- Min/max execution time range
- Estimated time to completion
- Comparison between commands (when benchmarking multiple commands)

## CSV Output

You can export benchmark results to a CSV file for further analysis:

```bash
cmdperf --csv=results.csv "sleep 0.1" "sleep 0.2"
```

## Markdown Output

You can export benchmark results to a Markdown file for documentation or sharing:

```bash
cmdperf --markdown=results.md "sleep 0.1" "sleep 0.2"
```

## License

MIT
