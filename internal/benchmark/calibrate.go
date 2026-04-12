package benchmark

import "time"

// TimerOverhead measures the typical cost of a single time.Since call.
// Useful context when interpreting sub-microsecond durations.
func TimerOverhead() time.Duration {
	const samples = 1000
	var total time.Duration
	for i := 0; i < samples; i++ {
		start := time.Now()
		total += time.Since(start)
	}
	return total / samples
}
