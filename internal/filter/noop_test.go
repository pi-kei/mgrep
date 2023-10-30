package filter

import (
	"testing"

	"github.com/pi-kei/mgrep/internal/base"
)

func TestNoopFilter_SkipDirEntry(t *testing.T) {
	filter := NewNoopFilter()
	
	skip := filter.SkipDirEntry(base.DirEntry{IsDir: true})
	if skip {
		t.Error("Returned true")
	}
}

func TestNoopFilter_SkipFileEntry(t *testing.T) {
	filter := NewNoopFilter()
	
	skip := filter.SkipFileEntry(base.DirEntry{IsDir: false})
	if skip {
		t.Error("Returned true")
	}
}

func TestNoopFilter_SkipSearchResult(t *testing.T) {
	filter := NewNoopFilter()
	
	skip := filter.SkipSearchResult(base.SearchResult{})
	if skip {
		t.Error("Returned true")
	}
}