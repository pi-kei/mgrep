package sink

import (
	"fmt"
	"io"

	"github.com/pi-kei/mgrep/internal/base"
)

type Writer struct {
	writer io.Writer
	format string
	getValues func(result base.SearchResult) []any
}

// Sink that writes formatted strings to a specified writer.
// Uses default format.
// Not thread-safe.
func NewWriterSink(writer io.Writer) base.Sink {
	return &Writer{writer, DefaultFormat, DefaultGetValues}
}

// Sink that writes formatted strings to a specified writer.
// Uses specified format.
// Not thread-safe.
func NewCustomWriterSink(writer io.Writer, format string, getValues func(result base.SearchResult) []any) base.Sink {
	return &Writer{writer, format, getValues}
}

func (w *Writer) HandleResult(result base.SearchResult) {
	fmt.Fprintf(w.writer, w.format, w.getValues(result)...)
}