package benchmark

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/miklosn/cmdperf/internal/command"
)

const (
	// DefaultBatchSize is the number of results to process in a batch to reduce lock contention
	DefaultBatchSize = 10

	// MinUpdateInterval is the minimum time between progress updates
	MinUpdateInterval = 100 * time.Millisecond

	// MaxRecentResults is the maximum number of recent results to store
	MaxRecentResults = 1000
)

// Options represents benchmark configuration options
type Options struct {
	// Number of iterations to run for each command
	Iterations int

	// Number of parallel workers for each command
	Parallelism int

	// Timeout for each command execution
	Timeout time.Duration

	// Total benchmark duration (overrides Iterations if set)
	Duration time.Duration
}

// CommandStats holds statistics for a single command
type CommandStats struct {
	// Command that was benchmarked
	Command *command.Command

	// Store only a fixed number of recent results to prevent unbounded memory growth
	RecentResults []*command.Result

	// Running statistics
	TotalRuns      int
	SuccessfulRuns int
	ErrorCount     int

	// Exit code tracking
	ExitCodes map[int]int // Maps exit code to count

	// Summary statistics
	Min, Max, Mean, Median, StdDev time.Duration

	// Running sum for incremental mean calculation
	RunningSum time.Duration

	// Representative sample of durations for median calculation
	MedianSamples []time.Duration

	// Throughput in operations per second
	Throughput float64

	// Timestamps for throughput calculation
	FirstStartTime time.Time
	LastEndTime    time.Time
}

// Runner coordinates the benchmark execution
type Runner struct {
	Options  Options
	Commands []*command.Command
	Results  []*CommandStats

	wg               sync.WaitGroup
	statsMutex       sync.Mutex
	progressCallback func(stats []*CommandStats, complete bool)
	eventHandler     func(event interface{})
}

// NewRunner creates a new benchmark runner with validation
func NewRunner(commands []*command.Command, options Options) (*Runner, error) {
	// Validate inputs
	if len(commands) == 0 {
		return nil, errors.New("benchmark: at least one command is required")
	}
	if options.Iterations <= 0 {
		return nil, errors.New("benchmark: iterations must be positive")
	}
	if options.Parallelism <= 0 {
		return nil, errors.New("benchmark: parallelism must be positive")
	}

	return &Runner{
		Options:  options,
		Commands: commands,
		Results:  make([]*CommandStats, len(commands)),
	}, nil
}

// SetProgressCallback sets a callback function to report progress
func (r *Runner) SetProgressCallback(callback func(stats []*CommandStats, complete bool)) {
	r.progressCallback = callback
}

// SetEventHandler sets a handler function for benchmark events
func (r *Runner) SetEventHandler(handler func(event interface{})) {
	r.eventHandler = handler
}

// contextCanceled is a helper function to check if context is canceled
func contextCanceled(ctx context.Context) bool {
	return ctx.Err() != nil
}

// Run executes the benchmark for all commands
func (runner *Runner) Run(ctx context.Context) {
	// Initialize results with pre-allocated capacity
	for i := range runner.Commands {
		runner.Results[i] = &CommandStats{
			Command:       runner.Commands[i],
			RecentResults: make([]*command.Result, 0, MaxRecentResults),
			MedianSamples: make([]time.Duration, 0, 1000),
			ExitCodes:     make(map[int]int), // Initialize exit code map
		}
	}

	// Emit benchmark started event
	if runner.eventHandler != nil {
		runner.eventHandler(map[string]interface{}{
			"type":     "benchmark_started",
			"commands": runner.Commands,
			"options":  runner.Options,
		})
	}

	// Create a context for the benchmark
	var benchCtx context.Context
	var benchCancel context.CancelFunc

	if runner.Options.Duration > 0 {
		// If duration is set, create a context with timeout
		benchCtx, benchCancel = context.WithTimeout(ctx, runner.Options.Duration)
	} else {
		// Otherwise use the parent context
		benchCtx, benchCancel = context.WithCancel(ctx)
	}
	defer benchCancel()

	// Start a timer to ensure regular UI updates even for slow commands
	updateTicker := time.NewTicker(500 * time.Millisecond)
	defer updateTicker.Stop()

	go func() {
		for {
			select {
			case <-updateTicker.C:
				// Send a progress update even if no commands have completed
				if runner.progressCallback != nil {
					runner.progressCallback(runner.Results, false)
				}
			case <-benchCtx.Done():
				return
			}
		}
	}()

	// Launch a goroutine for each command
	for cmdIndex, command := range runner.Commands {
		runner.wg.Add(1)
		go runner.runCommand(benchCtx, cmdIndex, command)
	}

	// Wait for all benchmarks to complete
	runner.wg.Wait()

	// Stop the update ticker
	benchCancel()

	// Calculate final statistics
	for _, stats := range runner.Results {
		// Recalculate standard deviation for final results
		calculateStdDev(stats)

		// Ensure median is up-to-date by sorting samples
		if len(stats.MedianSamples) > 0 {
			sort.Slice(stats.MedianSamples, func(i, j int) bool {
				return stats.MedianSamples[i] < stats.MedianSamples[j]
			})

			// Get median from sorted samples
			midIdx := len(stats.MedianSamples) / 2
			if len(stats.MedianSamples)%2 == 1 {
				stats.Median = stats.MedianSamples[midIdx]
			} else {
				stats.Median = (stats.MedianSamples[midIdx-1] + stats.MedianSamples[midIdx]) / 2
			}
		}
	}

	// Final progress report
	if runner.progressCallback != nil {
		runner.progressCallback(runner.Results, true)
	}

	// Emit benchmark completed event
	if runner.eventHandler != nil {
		runner.eventHandler(map[string]interface{}{
			"type":    "benchmark_completed",
			"results": runner.Results,
		})
	}
}

// emitCommandStarted emits a command started event
func (runner *Runner) emitCommandStarted(index int, cmd *command.Command) {
	if runner.eventHandler == nil {
		return
	}

	runner.eventHandler(map[string]interface{}{
		"type":          "command_started",
		"command_index": index,
		"command":       cmd,
	})
}

// emitCommandProgress emits a command progress event
func (runner *Runner) emitCommandProgress(index int, cmd *command.Command, result *command.Result, completed, total int) {
	if runner.eventHandler == nil {
		return
	}

	progress := float64(completed) / float64(total)
	runner.statsMutex.Lock()
	statsCopy := *runner.Results[index]
	runner.statsMutex.Unlock()

	// Create event data
	eventData := map[string]interface{}{
		"type":          "command_progress",
		"command_index": index,
		"command":       cmd,
		"stats":         statsCopy,
		"progress":      progress,
		"completed":     completed,
		"total":         total,
	}

	// Only include result if not nil
	if result != nil {
		eventData["result"] = result
	}

	runner.eventHandler(eventData)
}

// emitCommandCompleted emits a command completed event
func (runner *Runner) emitCommandCompleted(index int, cmd *command.Command) {
	if runner.eventHandler == nil {
		return
	}

	runner.statsMutex.Lock()
	statsCopy := *runner.Results[index]
	runner.statsMutex.Unlock()

	runner.eventHandler(map[string]interface{}{
		"type":          "command_completed",
		"command_index": index,
		"command":       cmd,
		"stats":         statsCopy,
	})
}

// processBatch processes a batch of results for a command
func (runner *Runner) processBatch(cmdIndex int, resultBatch []*command.Result) {
	if len(resultBatch) == 0 {
		return
	}

	// Process each result individually to ensure immediate updates
	for _, result := range resultBatch {
		runner.statsMutex.Lock()

		cmdStats := runner.Results[cmdIndex]

		// Add to recent results buffer
		addResultToRecentResults(cmdStats, result)

		// Update statistics incrementally
		updateStatsIncrementally(cmdStats, result)

		// Unlock before calling the callback to avoid deadlocks
		runner.statsMutex.Unlock()

		// Report progress after EACH result to show progress as soon as possible
		// This is especially important for the first few results
		if runner.progressCallback != nil {
			runner.progressCallback(runner.Results, false)
		}
	}
}

// addResultToRecentResults adds a result to the recent results buffer
func addResultToRecentResults(stats *CommandStats, result *command.Result) {
	// If we've reached the limit, make room by removing oldest result
	if len(stats.RecentResults) >= MaxRecentResults {
		// Shift elements to remove the oldest (index 0)
		copy(stats.RecentResults, stats.RecentResults[1:])
		stats.RecentResults = stats.RecentResults[:MaxRecentResults-1]
	}

	// Add the new result
	stats.RecentResults = append(stats.RecentResults, result)
}

// updateStatsIncrementally updates statistics incrementally with a new result
func updateStatsIncrementally(stats *CommandStats, newResult *command.Result) {
	// Update total runs counter
	stats.TotalRuns++

	// Track exit code
	stats.ExitCodes[newResult.ExitCode]++

	// Check for errors
	if newResult.Error != nil {
		// For duration-based benchmarks, don't count context cancellation as an error
		isDurationTimeout := stats.Command != nil &&
			stats.Command.Parallelism > 0 && // Just check if it's a valid command
			stats.Command.Timeout > 0 && // Check if command has a timeout set
			newResult.ContextCancelled

		if !isDurationTimeout {
			stats.ErrorCount++
		}

		// For timeouts, we might want to skip timing stats
		if newResult.TimedOut {
			// Skip timing stats for timeouts
			return
		}
		// For other errors, continue to update timing stats
	}

	// Update successful runs counter (commands that executed, even with errors)
	stats.SuccessfulRuns++

	// Get the duration
	duration := newResult.Duration

	// Update min/max
	if stats.SuccessfulRuns == 1 || duration < stats.Min {
		stats.Min = duration
	}
	if stats.SuccessfulRuns == 1 || duration > stats.Max {
		stats.Max = duration
	}

	// Update running sum for mean calculation
	stats.RunningSum += duration
	stats.Mean = stats.RunningSum / time.Duration(stats.SuccessfulRuns)

	// Update median samples (reservoir sampling)
	updateMedianSamples(stats, duration)

	// Update throughput calculation
	updateThroughputStats(stats, newResult)

	// Periodically recalculate standard deviation (more expensive)
	if stats.SuccessfulRuns%100 == 0 || stats.SuccessfulRuns <= 10 {
		calculateStdDev(stats)
	}
}

// Helper functions for incremental statistics calculation

// Constants for median calculation
const (
	MaxMedianSamples            = 1000
	MedianResortInterval        = 250
	MedianInitialResortInterval = 50
)

// updateMedianSamples updates median samples using reservoir sampling
func updateMedianSamples(stats *CommandStats, duration time.Duration) {
	if len(stats.MedianSamples) < MaxMedianSamples {
		// Still building initial sample set
		stats.MedianSamples = append(stats.MedianSamples, duration)

		// If we just filled the sample set, sort it
		if len(stats.MedianSamples) == MaxMedianSamples {
			sortMedianSamples(stats)
		}
	} else {
		// Use reservoir sampling to maintain a representative sample
		if rand.Intn(stats.SuccessfulRuns) < MaxMedianSamples {
			// Replace a random sample
			sampleIndex := rand.Intn(MaxMedianSamples)
			stats.MedianSamples[sampleIndex] = duration

			// Resort the samples periodically
			if stats.SuccessfulRuns%MedianResortInterval == 0 {
				sortMedianSamples(stats)
			}
		}
	}

	// Calculate median from samples if we have enough
	if len(stats.MedianSamples) > 0 {
		// Sort only periodically to reduce CPU usage
		if stats.SuccessfulRuns%MedianInitialResortInterval == 0 || stats.SuccessfulRuns <= 5 {
			sortMedianSamples(stats)
		}

		// Get median from sorted samples
		updateMedianValue(stats)
	}
}

// sortMedianSamples sorts the median samples in ascending order
func sortMedianSamples(stats *CommandStats) {
	sort.Slice(stats.MedianSamples, func(i, j int) bool {
		return stats.MedianSamples[i] < stats.MedianSamples[j]
	})
}

// updateMedianValue calculates the median value from sorted samples
func updateMedianValue(stats *CommandStats) {
	midIdx := len(stats.MedianSamples) / 2
	if len(stats.MedianSamples)%2 == 1 {
		stats.Median = stats.MedianSamples[midIdx]
	} else {
		stats.Median = (stats.MedianSamples[midIdx-1] + stats.MedianSamples[midIdx]) / 2
	}
}

// updateThroughputStats updates throughput statistics
func updateThroughputStats(stats *CommandStats, result *command.Result) {
	// Update first start time and last end time
	endTime := result.StartTime.Add(result.Duration)

	if stats.SuccessfulRuns == 1 {
		stats.FirstStartTime = result.StartTime
		stats.LastEndTime = endTime
	} else {
		if result.StartTime.Before(stats.FirstStartTime) {
			stats.FirstStartTime = result.StartTime
		}
		if endTime.After(stats.LastEndTime) {
			stats.LastEndTime = endTime
		}
	}

	// Calculate throughput
	totalTime := stats.LastEndTime.Sub(stats.FirstStartTime)
	if totalTime > 0 {
		stats.Throughput = float64(stats.SuccessfulRuns) / totalTime.Seconds()
	}
}

// calculateStdDev calculates standard deviation from recent results
func calculateStdDev(stats *CommandStats) {
	// If we don't have enough samples, use all recent results
	if len(stats.RecentResults) < 2 {
		return
	}

	// Calculate variance using Welford's online algorithm
	var m2 float64
	mean := float64(stats.Mean.Nanoseconds())
	count := 0

	for _, result := range stats.RecentResults {
		if result.Error != nil {
			continue
		}

		delta := float64(result.Duration.Nanoseconds()) - mean
		m2 += delta * delta
		count++
	}

	if count > 1 {
		variance := m2 / float64(count)
		stats.StdDev = time.Duration(math.Sqrt(variance))
	}
}
