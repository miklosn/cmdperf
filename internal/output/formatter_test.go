package output

import (
	"testing"
	"time"
)

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		expected string
	}{
		{500 * time.Nanosecond, "500.00 ns"},
		{1500 * time.Nanosecond, "1.50 µs"},
		{1500 * time.Microsecond, "1.50 ms"},
		{1500 * time.Millisecond, "1.50 s"},
	}

	for _, test := range tests {
		result := FormatDuration(test.duration)
		if result != test.expected {
			t.Errorf("FormatDuration(%v) = %s, expected %s",
				test.duration, result, test.expected)
		}
	}
}

func TestFormatThroughput(t *testing.T) {
	tests := []struct {
		throughput float64
		expected   string
	}{
		{0, "-"},
		{-1, "-"},
		{5, "5.00 /s"},
		{500, "500.00 /s"},
		{1500, "1.50 K/s"},
		{1500000, "1.50 M/s"},
		{0.5, "0.50 /s"},
		{0.01, "0.60 /min"},
		{0.0001, "0.36 /hr"},
	}

	for _, test := range tests {
		result := FormatThroughput(test.throughput)
		if result != test.expected {
			t.Errorf("FormatThroughput(%f) = %s, expected %s",
				test.throughput, result, test.expected)
		}
	}
}
