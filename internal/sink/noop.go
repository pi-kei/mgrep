package sink

import (
	"github.com/pi-kei/mgrep/internal/base"
)

type Noop struct {}

func NewNoopSink() base.Sink {
	return &Noop{}
}

func (n *Noop) HandleResult(result base.SearchResult) {
	// noop
}