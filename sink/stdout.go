package sink

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/pi-kei/mgrep/base"
)

type Stdout struct {
	highlight func(a ...interface{}) string
}

func NewStdoutSink() base.Sink {
	var highlight = color.New(color.Bold, color.FgHiYellow).SprintFunc()
	return &Stdout{highlight}
}

func (s *Stdout) HandleResult(result base.SearchResult) {
	startPart := result.Line[0:result.StartIndex]
	resultPart := s.highlight(result.Line[result.StartIndex:result.EndIndex])
	endPart := result.Line[result.EndIndex:]
	fmt.Printf("%s[%v,%v]:%s%s%s\n", result.Path, result.LineNumber, result.StartIndex+1, startPart, resultPart, endPart)
}