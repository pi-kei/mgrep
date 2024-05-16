package filter

import (
	"github.com/pi-kei/mgrep/internal/base"
)

type Noop struct{}

func NewNoop() base.Filter {
	return &Noop{}
}

func (n *Noop) SkipDirEntry(dirEntry base.DirEntry) bool {
	return false
}

func (n *Noop) SkipFileEntry(fileEntry base.DirEntry) bool {
	return false
}

func (n *Noop) SkipSearchResult(searchResult base.SearchResult) bool {
	return false
}
