package skipper

import (
	"github.com/pi-kei/mgrep/base"
)

type Noop struct{}

func NewNoopSkipper() base.Skipper {
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