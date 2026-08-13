//go:build linux || darwin

package benchmark

import "runtime"

// pinWorkerThread pins the calling goroutine to its OS thread so timing
// samples aren't perturbed by Go scheduler migrations. Measured as a win on
// Linux and macOS, but a regression on FreeBSD (see pin_thread_off.go).
func pinWorkerThread() func() {
	runtime.LockOSThread()
	return runtime.UnlockOSThread
}
