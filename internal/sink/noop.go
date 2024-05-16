package sink

import (
	"github.com/pi-kei/mgrep/internal/base"
)

type Noop struct{}

// Sink that does nothing.
// Thread-safe.
func NewNoop() base.Sink {
	return &Noop{}
}

func (n *Noop) HandleResult(result base.SearchResult) {
	// noop
}
