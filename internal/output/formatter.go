package output

import (
	"fmt"
	"time"
)

func FormatDuration(d time.Duration) string {
	if d < time.Microsecond {
		return fmt.Sprintf("%.2f ns", float64(d.Nanoseconds()))
	} else if d < time.Millisecond {
		return fmt.Sprintf("%.2f µs", float64(d.Nanoseconds())/1000)
	} else if d < time.Second {
		return fmt.Sprintf("%.2f ms", float64(d.Nanoseconds())/1000000)
	} else {
		return fmt.Sprintf("%.2f s", d.Seconds())
	}
}

func FormatThroughput(throughput float64) string {
	if throughput <= 0 {
		return "-"
	}
	if throughput >= 1000000 {
		return fmt.Sprintf("%.2f M/s", throughput/1000000)
	} else if throughput >= 1000 {
		return fmt.Sprintf("%.2f K/s", throughput/1000)
	} else if throughput >= 0.5 {
		return fmt.Sprintf("%.2f /s", throughput)
	} else if throughput >= 0.01 {
		return fmt.Sprintf("%.2f /min", throughput*60)
	} else {
		return fmt.Sprintf("%.2f /hr", throughput*3600)
	}
}
