package benchmark

import (
	"context"
	"math"
	"sync"
	"time"

	"github.com/miklosn/cmdperf/internal/command"
)

func (runner *Runner) runCommand(ctx context.Context, index int, cmd *command.Command) {
	defer runner.wg.Done()

	runner.emitCommandStarted(index, cmd)

	if contextCanceled(ctx) {
		return
	}

	workerCtx, workerCancel := context.WithCancel(ctx)
	defer workerCancel()

	// Determine if we're using duration-based or iteration-based benchmarking
	var workCh chan int
	var totalIterations int

	if runner.Options.Duration > 0 {
		// For duration-based benchmarking, use a buffered channel with a large capacity
		workCh = make(chan int, 1000)
		totalIterations = math.MaxInt32 // Use a very large number for progress calculation

		// Start a goroutine to feed the work channel until context is cancelled
		go func() {
			defer close(workCh)
			i := 0
			for {
				select {
				case workCh <- i:
					i++
				case <-workerCtx.Done():
					return
				}
			}
		}()
	} else {
		// For iteration-based benchmarking, use the existing approach
		workCh = make(chan int, runner.Options.Iterations)
		totalIterations = runner.Options.Iterations

		for i := 0; i < runner.Options.Iterations; i++ {
			select {
			case workCh <- i:
			case <-workerCtx.Done():
				close(workCh)
				return
			}
		}
		close(workCh)
	}

	resultCh := make(chan *command.Result, runner.Options.Parallelism*2)
	completedIterations := 0

	var workerWg sync.WaitGroup
	parallelismTokens := make(chan struct{}, runner.Options.Parallelism)
	for i := 0; i < runner.Options.Parallelism; i++ {
		parallelismTokens <- struct{}{}
	}

	for workerID := 0; workerID < runner.Options.Parallelism; workerID++ {
		workerWg.Add(1)
		go func(workerID int) {
			defer workerWg.Done()

			// Process work items assigned to this worker
			for range workCh {
				if contextCanceled(workerCtx) {
					return
				}

				select {
				case <-parallelismTokens:
				case <-workerCtx.Done():
					return
				}

				// Execute command
				result := cmd.Execute(workerCtx)
				// Check if the context was cancelled and set the flag
				if workerCtx.Err() != nil {
					result.ContextCancelled = true
				}

				select {
				case parallelismTokens <- struct{}{}:
				default:
				}

				select {
				case resultCh <- result:
				case <-workerCtx.Done():
					return
				}
			}
		}(workerID)
	}

	go func() {
		workerWg.Wait()
		close(resultCh)
	}()

	progressTicker := time.NewTicker(250 * time.Millisecond)
	defer progressTicker.Stop()

	go func() {
		for {
			select {
			case <-progressTicker.C:
				// Report progress periodically even if no results yet
				if runner.progressCallback != nil {
					// Pass the current command index and total iterations to show progress
					runner.emitCommandProgress(index, cmd, nil, completedIterations, totalIterations)
					runner.progressCallback(runner.Results, false)
				}
			case <-workerCtx.Done():
				return
			}
		}
	}()

	batchSize := DefaultBatchSize
	if runner.Options.Parallelism > 4 {
		batchSize = DefaultBatchSize * (runner.Options.Parallelism / 4)
		if batchSize > 50 {
			batchSize = 50
		}
	}

	resultBatch := make([]*command.Result, 0, batchSize)
	lastEventTime := time.Now()
	lastProgressTime := time.Now()

	for result := range resultCh {
		resultBatch = append(resultBatch, result)
		completedIterations++

		shouldProcessBatch := len(resultBatch) >= batchSize ||
			completedIterations == totalIterations ||
			contextCanceled(ctx)
		if completedIterations <= 5 || time.Since(lastProgressTime) >= 100*time.Millisecond {
			shouldProcessBatch = true
			lastProgressTime = time.Now()
		}

		if shouldProcessBatch && len(resultBatch) > 0 {
			runner.processBatch(index, resultBatch)

			if runner.progressCallback != nil {
				runner.progressCallback(runner.Results, false)
			}
			resultBatch = resultBatch[:0]
		}

		now := time.Now()
		shouldUpdate := now.Sub(lastEventTime) >= MinUpdateInterval
		if shouldUpdate || result.Error != nil || result.Duration > time.Second {
			runner.emitCommandProgress(index, cmd, result, completedIterations, totalIterations)
			lastEventTime = now
		}

		if contextCanceled(ctx) {
			workerCancel()
			break
		}
	}

	runner.emitCommandCompleted(index, cmd)
}
