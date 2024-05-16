package filter

import (
	"testing"

	"github.com/pi-kei/mgrep/internal/base"
)

func TestConfigurableFilter_SkipDirEntry(t *testing.T) {
	calledTimes := 0
	skipDirEntry := func(entry base.DirEntry) bool {
		calledTimes++
		return true
	}
	skipFileEntry := func(entry base.DirEntry) bool {
		return false
	}
	skipSearchResult := func(result base.SearchResult) bool {
		return false
	}
	filter := NewConfigurable(skipDirEntry, skipFileEntry, skipSearchResult)

	skip := filter.SkipDirEntry(base.DirEntry{IsDir: true})
	if !skip {
		t.Error("Returned false")
	}
	if calledTimes != 1 {
		t.Errorf("Called times %v", calledTimes)
	}
}

func TestConfigurableFilter_SkipFileEntry(t *testing.T) {
	calledTimes := 0
	skipDirEntry := func(entry base.DirEntry) bool {
		return false
	}
	skipFileEntry := func(entry base.DirEntry) bool {
		calledTimes++
		return true
	}
	skipSearchResult := func(result base.SearchResult) bool {
		return false
	}
	filter := NewConfigurable(skipDirEntry, skipFileEntry, skipSearchResult)

	skip := filter.SkipFileEntry(base.DirEntry{IsDir: false})
	if !skip {
		t.Error("Returned false")
	}
	if calledTimes != 1 {
		t.Errorf("Called times %v", calledTimes)
	}
}

func TestConfigurableFilter_SkipSearchResult(t *testing.T) {
	calledTimes := 0
	skipDirEntry := func(entry base.DirEntry) bool {
		return false
	}
	skipFileEntry := func(entry base.DirEntry) bool {
		return false
	}
	skipSearchResult := func(result base.SearchResult) bool {
		calledTimes++
		return true
	}
	filter := NewConfigurable(skipDirEntry, skipFileEntry, skipSearchResult)

	skip := filter.SkipSearchResult(base.SearchResult{})
	if !skip {
		t.Error("Returned false")
	}
	if calledTimes != 1 {
		t.Errorf("Called times %v", calledTimes)
	}
}
