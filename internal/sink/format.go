package sink

import (
	"unicode/utf8"

	"github.com/fatih/color"
	"github.com/pi-kei/mgrep/internal/base"
)

var DefaultFormat = "%s[%d,%d]:%s%s%s\n"
var highlight = color.New(color.Bold, color.FgHiYellow).SprintFunc()

func DefaultGetValues(result base.SearchResult) []any {
	startPart := result.Line[0:result.StartIndex]
	resultPart := highlight(result.Line[result.StartIndex:result.EndIndex])
	endPart := result.Line[result.EndIndex:]
	return []any{result.Path, result.LineNumber, utf8.RuneCountInString(startPart) + 1, startPart, resultPart, endPart}
}
