package sink

import (
	"fmt"
	"io"

	"github.com/fatih/color"
	"github.com/pi-kei/mgrep/base"
)

type Writer struct {
	writer io.Writer
	highlight func(a ...interface{}) string
}

func NewWriterSink(writer io.Writer) base.Sink {
	var highlight = color.New(color.Bold, color.FgHiYellow).SprintFunc()
	return &Writer{writer, highlight}
}

func (w *Writer) HandleResult(result base.SearchResult) {
	startPart := result.Line[0:result.StartIndex]
	resultPart := w.highlight(result.Line[result.StartIndex:result.EndIndex])
	endPart := result.Line[result.EndIndex:]
	fmt.Fprintf(w.writer, "%s[%v,%v]:%s%s%s\n", result.Path, result.LineNumber, result.StartIndex+1, startPart, resultPart, endPart)
}