package sink

import (
	"testing"

	"github.com/pi-kei/mgrep/internal/base"
)

func TestNoopSink_HandleResult(t *testing.T) {
	sink := NewNoopSink()

	sink.HandleResult(base.SearchResult{Path: "a/b/c.txt", LineNumber: 1, StartIndex: 2, EndIndex: 5, Line: "test test"})
}