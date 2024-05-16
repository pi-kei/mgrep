package sink

import (
	"strings"
	"testing"

	"github.com/pi-kei/mgrep/internal/base"
)

func TestWriterSink_HandleResult(t *testing.T) {
	var sb strings.Builder
	sink := NewWriter(&sb)

	sink.HandleResult(base.SearchResult{Path: "a/b/c.txt", LineNumber: 1, StartIndex: 2, EndIndex: 5, Line: "test test"})
	out := sb.String()
	if out != "a/b/c.txt[1,3]:test test\n" {
		t.Errorf("Invalid output: %s", out)
	}
}

func TestCustomWriterSink_HandleResult(t *testing.T) {
	var sb strings.Builder
	calledTimes := 0
	sink := NewWriter(&sb, WithWriterFormat("%s %v %v %v %s"), WithWriterGetValues(func(result base.SearchResult) []any {
		calledTimes++
		return []any{result.Path, result.LineNumber, result.StartIndex, result.EndIndex, result.Line}
	}))

	sink.HandleResult(base.SearchResult{Path: "a/b/c.txt", LineNumber: 1, StartIndex: 2, EndIndex: 5, Line: "test test"})
	out := sb.String()
	if out != "a/b/c.txt 1 2 5 test test" {
		t.Errorf("Invalid output: %s", out)
	}
	if calledTimes != 1 {
		t.Errorf("Called times %v", calledTimes)
	}
}
