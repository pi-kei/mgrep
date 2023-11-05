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

func NewWriterSink(writer io.Writer) base.Sink {
	return &Writer{writer, DefaultFormat, DefaultGetValues}
}

func NewCustomWriterSink(writer io.Writer, format string, getValues func(result base.SearchResult) []any) base.Sink {
	return &Writer{writer, format, getValues}
}

func (w *Writer) HandleResult(result base.SearchResult) {
	fmt.Fprintf(w.writer, w.format, w.getValues(result)...)
}