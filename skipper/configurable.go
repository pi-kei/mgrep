package skipper

import (
	"github.com/pi-kei/mgrep/base"
)

type Configurable struct{
	skipDirEntryFn func(dirEntry base.DirEntry) bool
	skipFileEntryFn func(fileEntry base.DirEntry) bool
	skipSearchResultFn func(searchResult base.SearchResult) bool
}

func NewConfigurableSkipper(
	skipDirEntryFn func(dirEntry base.DirEntry) bool,
	skipFileEntryFn func(fileEntry base.DirEntry) bool,
	skipSearchResultFn func(searchResult base.SearchResult) bool,
) base.Skipper {
	return &Configurable{skipDirEntryFn, skipFileEntryFn, skipSearchResultFn}
}

func (c *Configurable) SkipDirEntry(dirEntry base.DirEntry) bool {
	return c.skipDirEntryFn(dirEntry)
}

func (c *Configurable) SkipFileEntry(fileEntry base.DirEntry) bool {
	return c.skipFileEntryFn(fileEntry)
}

func (c *Configurable) SkipSearchResult(searchResult base.SearchResult) bool {
	return c.skipSearchResultFn(searchResult)
}