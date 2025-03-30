package output

import (
	"fmt"
	"io"

	"github.com/miklosn/cmdperf/internal/benchmark"
)

type Writer interface {
	Write(w io.Writer, stats []*benchmark.CommandStats) error
}

func GetWriter(format string) (Writer, error) {
	switch format {
	case "csv":
		return &CSVWriter{}, nil
	case "markdown":
		return &MarkdownWriter{}, nil
	case "terminal":
		return &TerminalWriter{}, nil
	default:
		return nil, fmt.Errorf("unsupported output format: %s", format)
	}
}
