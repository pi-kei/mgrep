package sink

import (
	"fmt"
	"io"

	"github.com/fatih/color"
	"github.com/pi-kei/mgrep/base"
)

type Writer struct {
	writer io.Writer
	format string
	getValues func(result base.SearchResult) []any
}

func NewWriterSink(writer io.Writer) base.Sink {
	var highlight = color.New(color.Bold, color.FgHiYellow).SprintFunc()
	return &Writer{writer, "%s[%v,%v]:%s%s%s\n", func(result base.SearchResult) []any {
		startPart := result.Line[0:result.StartIndex]
		resultPart := highlight(result.Line[result.StartIndex:result.EndIndex])
		endPart := result.Line[result.EndIndex:]
		return []any{result.Path, result.LineNumber, result.StartIndex+1, startPart, resultPart, endPart}
	}}
}

func NewCustomWriterSink(writer io.Writer, format string, getValues func(result base.SearchResult) []any) base.Sink {
	return &Writer{writer, format, getValues}
}

func (w *Writer) HandleResult(result base.SearchResult) {
	fmt.Fprintf(w.writer, w.format, w.getValues(result)...)
}