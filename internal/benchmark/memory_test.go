package benchmark

import (
	"context"
	"runtime"
	"testing"
	"time"

	"github.com/miklosn/cmdperf/internal/command"
)

func TestMemoryUsage(t *testing.T) {
	// Skip if short test
	if testing.Short() {
		t.Skip("Skipping memory test in short mode")
	}

	// Create a simple command that does nothing
	cmd := &command.Command{
		Command:     "true",
		Raw:         "true",
		DirectExec:  true,
		Args:        []string{},
		Parallelism: 1,
	}

	// Options for a long run
	iterations := 50000
	options := Options{
		Iterations:  iterations,
		Parallelism: 50,
		Timeout:     1 * time.Second,
	}

	runner, err := NewRunner([]*command.Command{cmd}, options)
	if err != nil {
		t.Fatalf("Failed to create runner: %v", err)
	}

	// Record initial memory
	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	// Run benchmark
	ctx := context.Background()
	runner.Run(ctx)

	// Record final memory
	runtime.GC()
	runtime.ReadMemStats(&m2)

	// Calculate difference
	diff := m2.Alloc - m1.Alloc
	diffMB := float64(diff) / 1024 / 1024

	t.Logf("Initial Alloc: %v MB", float64(m1.Alloc)/1024/1024)
	t.Logf("Final Alloc: %v MB", float64(m2.Alloc)/1024/1024)
	t.Logf("Memory Diff: %.2f MB", diffMB)
	t.Logf("Mallocs Diff: %d", m2.Mallocs-m1.Mallocs)

	// Check if we have excessive remaining allocations.
	// Result struct is small, but if we have 100,000 iterations and they are not pooled/GC'd effectively,
	// we might see high memory usage.
	// However, since we expect GC to clean up if we don't reference them, high allocs + GC should mean low *current* Alloc.
	// But if we leak, Alloc will be high.

	// If the leak is "real" (references kept), m2.Alloc will be high.
	// If the issue is just "high churn", m2.Mallocs will be high, but m2.Alloc might be low (because GC cleaned up).
	// The Review says "Memory grows unbounded", implying Alloc grows.

	// Let's assert on Alloc.
	// 100,000 results * sizeof(Result). Result has pointers, time, etc. maybe 100-200 bytes.
	// 100,000 * 200 = 20 MB.
	// If we only keep 1000, we should be much lower.

	if diffMB > 10 {
		t.Errorf("Memory usage increased by %.2f MB, indicating a potential leak", diffMB)
	}
}
