# cmdperf: Command Performance Benchmarking Tool

## Overview

`cmdperf` is a command-line benchmarking tool written in Go that allows users to measure and compare the performance of shell commands. It provides real-time progress tracking, colorful terminal output, and support for various output formats.

## Core Features

1. Command benchmarking with iteration or duration-based execution
2. Parallel execution support
3. Real-time terminal UI with progress tracking
4. Multiple color schemes
5. CSV and Markdown output formats
6. Direct command execution mode (no shell)
7. Profiling support
8. Graceful cancellation handling


## Architecture

### High-Level Components

1. **CLI Interface**
   - Uses Kong for argument parsing and command handling
   - Validates input parameters
   - Forwards parameters to the benchmark runner

2. **Benchmark Runner**
   - Coordinates the execution of benchmark runs
   - Manages parallelism through goroutines
   - Collects timing results from command executions
   - Calculates statistics from timing data

3. **Command Executor**
   - Prepares shell commands for execution
   - Handles shell environment setup
   - Executes commands and captures timing information
   - Manages timeouts and command termination

4. **Results Display**
   - Formats and displays benchmark results
   - Compares multiple command results when applicable

### Concurrency Model: Independent Parallel Execution

The concurrency model is based on independent execution of commands with configurable parallelism. Each command runs with its own pool of worker goroutines, and all commands execute simultaneously.

```
                     ┌──────────────────┐
                     │   Main Program   │
                     └────────┬─────────┘
                              │
                              ▼
              ┌─────────────────────────────┐
              │      Benchmark Runner       │
              └──┬─────────┬─────────┬──────┘
                 │         │         │ (launch all commands simultaneously)
    ┌────────────▼─┐ ┌─────▼─────┐ ┌─▼──────────┐
    │ Command 1    │ │ Command 2 │ │ Command 3  │
    │ Executor     │ │ Executor  │ │ Executor   │
    └──────┬───────┘ └─────┬─────┘ └─────┬──────┘
           │               │             │
           ▼               ▼             ▼
┌───────────────────┐     ...           ...
│  Parallelism Pool │
│  for Command 1    │
└─┬───────┬─────────┘
  │       │
┌─▼─┐   ┌─▼─┐
│W 1│   │W 2│ ... (Workers per command = parallelism value)
└─┬─┘   └─┬─┘
  │       │
  ▼       ▼
Runs    Runs     (Each worker runs iterations/parallelism runs)

                 ┌─────────────────┐
                 │ Results         │
                 │ Aggregator      │
                 └────────┬────────┘
                          │
                          ▼
                 ┌─────────────────┐
                 │  Statistics     │
                 │  Processor      │
                 └────────┬────────┘
                          │
                          ▼
                 ┌─────────────────┐
                 │  Output         │
                 │  Formatter      │
                 └─────────────────┘
```

With this model:
- Total concurrent processes = Number of commands × Parallelism setting
- Each command runs independently at its own pace
- Results are collected and displayed as they become available
- Final comparison is shown when all commands complete

### Project Structure

```
cmdperf/
├── cmd/
│   └── cmdperf/
│       └── main.go                 # Entry point for the application
├── internal/
│   ├── benchmark/        # Core benchmarking functionality
│   │   ├── benchmark.go            # Core statistics and data structures
│   │   └── benchmark_runner.go     # Benchmark execution coordination
│   ├── command/          # Command handling
│   │   └── command.go              # Command preparation and execution
│   ├── output/           # Output formatting
│   │   ├── csv.go                  # CSV output formatter
│   │   ├── formatter.go            # Common formatting utilities
│   │   ├── markdown.go             # Markdown output formatter
│   │   ├── output.go               # Output interface definitions
│   │   └── terminal.go             # Terminal output formatter
│   └── ui/               # User interface components
│       ├── colorscheme/  # Terminal color schemes
│       │   ├── colorscheme.go      # Color scheme handling
│       │   └── schemes.go          # Color scheme definitions
│       ├── inline.go               # Inline terminal UI
│       └── util.go                 # UI utilities
├── go.mod                # Module definition
├── go.sum                # Dependency checksums
├── LICENSE               # Project license
├── README.md            # Project documentation
└── DESIGN.md            # This design document
```

## Implementation Guidelines

### Error Handling

Error handling should be comprehensive but not overly complicated:

- Distinguish between command execution errors and internal errors
- Provide meaningful error messages to the user
- Handle timeouts gracefully
- Clean up resources properly when errors occur

### Memory Management

- Pre-allocate slices where possible
- Limit output capture for verbose mode
- Use buffered channels appropriately
- Clear resources after each command execution

### Concurrency Management

- Calculate proper work distribution among goroutines
- Use contexts for timeout and cancellation
- Synchronize access to shared data with mutexes
- Use wait groups to coordinate goroutine completion

### Signal Handling and Graceful Shutdown

```go
// Setup signal handling for graceful shutdown
func setupSignalHandling(ctx context.Context, cancel context.CancelFunc) {
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
    
    go func() {
        select {
        case <-sigCh:
            fmt.Println("\nReceived interrupt signal, shutting down...")
            cancel() // Cancel the context to stop ongoing benchmarks
        case <-ctx.Done():
            // Context was canceled elsewhere, do nothing
        }
    }()
}
```

## Testing Strategy

### Unit Tests

- Test command execution in isolation
- Mock external commands for deterministic testing
- Test statistics calculations with known inputs

### Integration Tests

- Test with real shell commands
- Verify parallelism behavior
- Test with different shells and command types

### Performance Tests

- Measure overhead of the tool itself
- Verify resource usage
- Test with varying parallelism levels

## Future Extensions (out of scope for initial implementation)

- Statistical analysis (outlier detection, significance testing)
- Persistent history of benchmark results
- Export to HTML reports with interactive charts
- Comparison with previous benchmark runs
- Optional resource usage tracking where platform-specific APIs allow
- JSON/CSV output formats
- ASCII histogram visualization
