package skipper

import "github.com/pi-kei/mgrep/base"

type Configurable struct{}

func NewConfigurableSkipper() base.Skipper {
	return &Configurable{}
}

func (c *Configurable) SkipDirEntry(dirEntry base.DirEntry, options base.SearchOptions) bool {
	return dirEntry.Depth > options.MaxDepth
}

func (c *Configurable) SkipFileEntry(fileEntry base.DirEntry, options base.SearchOptions) bool {
	if fileEntry.Size == 0 {
		// nothing to search
		return true
	}
	if fileEntry.Size > options.MaxSize {
		// skip file because of options
		return true
	}
	if options.Include != nil && !options.Include.MatchString(fileEntry.Path) {
		// skip file because of options
		return true
	}
	if options.Exclude != nil && options.Exclude.MatchString(fileEntry.Path) {
		// skip file because of options
		return true
	}
	return false
}

func (c *Configurable) SkipSearchResult(searchResult base.SearchResult, options base.SearchOptions) bool {
	return len(searchResult.Line) > options.MaxLength
}