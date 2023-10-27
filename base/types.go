package base

import (
	"context"
	"io"
	"regexp"
	"time"
)

// Dir entry
type DirEntry struct {
	Path     string      // path to entry
	Depth    int         // recursion depth
	IsDir    bool        // whether the entry describes a directory
	Size     int64       // size of a file in bytes
	ModTime  time.Time   // modification time
}

// Represents a single match
type SearchResult struct {
	Path       string // path to file
	LineNumber int    // line number 1-based
	StartIndex int    // start index of a match 0-based
	EndIndex   int    // end index (exclusive) of a match 0-based
	Line       string // full line that has a match
}

// Reads one entry
type Reader interface {
	// Opens entry to read its content.
	// Entry must be a file
	OpenFile(fileEntry DirEntry) (interface { io.Reader; io.Closer }, error)
	// Reads child entries from specified entry.
	// Entry must be a directory
	ReadDir(dirEntry DirEntry) ([]DirEntry, error)
	// Reads root entry from specified name.
	// Entry can be a file or directory
	ReadRootEntry(name string) (DirEntry, error)
}

// Scans for matches
type Scanner interface {
	// Scans a file and calls a callback on each match
	ScanFile(fileEntry DirEntry, searchRegexp *regexp.Regexp, callback func(SearchResult) error) error
	// Scans directories starting at the specified root path and calls a callback on each found file
	ScanDirs(rootPath string, callback func(DirEntry) error) error
}

// Checks if skip is needed
type Filter interface {
	// Checks if skip is needed for directory.
	// It means that child entries must not be read
	SkipDirEntry(dirEntry DirEntry) bool
	// Checks if skip is needed for file.
	// It means that file content must not be read
	SkipFileEntry(fileEntry DirEntry) bool
	// Checks if skip is needed for search result.
	// It means that result must be ignored
	SkipSearchResult(dirEntry SearchResult) bool
}

// Handles search results
type Sink interface {
	// Handles search result
	HandleResult(result SearchResult)
}

// Performs search
type Searcher interface {
	// Starts search
	Search(rootPath string, searchRegexp *regexp.Regexp, ctx context.Context)
}