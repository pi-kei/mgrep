package sink

import (
	"fmt"
	"io"

	"github.com/pi-kei/mgrep/internal/base"
)

type Writer struct {
	writer    io.Writer
	format    string
	getValues func(result base.SearchResult) []any
}

type WriterOption func(*Writer)

func WithWriterFormat(format string) WriterOption {
	return func(w *Writer) {
		w.format = format
	}
}

func WithWriterGetValues(getValues func(result base.SearchResult) []any) WriterOption {
	return func(w *Writer) {
		w.getValues = getValues
	}
}

// Sink that writes formatted strings to a specified writer.
// Not thread-safe.
func NewWriter(writer io.Writer, options ...WriterOption) base.Sink {
	sink := Writer{writer, DefaultFormat, DefaultGetValues}
	for _, option := range options {
		option(&sink)
	}
	return &sink
}

func (w *Writer) HandleResult(result base.SearchResult) {
	fmt.Fprintf(w.writer, w.format, w.getValues(result)...)
}
