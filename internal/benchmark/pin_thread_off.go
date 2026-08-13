//go:build !linux && !darwin

package benchmark

// pinWorkerThread is a no-op on platforms where OS-thread pinning measured
// as a regression: on FreeBSD it added 5-15% mean and 25-107% stddev in CI
// benchmarks (cmdperf PR #14 investigation).
func pinWorkerThread() func() {
	return func() {}
}
