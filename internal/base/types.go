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

// Generic iterator
type Iterator[T any] interface {
	// Goes to next element.
	// Returns false when there are no more elements or error occured.
	// If false is returned then any additional call to Next() would do nothing
	Next() bool
	// Returns current value.
	// Call after Next().
	// If Next() was not called yet then Value() returns default value of T.
	// If Next() returned false then Value() returns the same value as before 
	Value() T
	// Returns last occured error or nil
	Err() error
}

// Reads one entry
type Reader interface {
	// Opens entry to read its content.
	// Entry must be a file
	OpenFile(fileEntry DirEntry) (io.ReadCloser, error)
	// Reads child entries from specified entry.
	// Entry must be a directory.
	// Returns iterator and error. Iterator can generate valid entries even if error is not nil
	ReadDir(dirEntry DirEntry) (Iterator[DirEntry], error)
	// Reads root entry from specified name.
	// Entry can be a file or directory
	ReadRootEntry(name string, depth int) (DirEntry, error)
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
	SkipSearchResult(searchResult SearchResult) bool
}

// Scans for matches
type Scanner interface {
	// Returns reader
	GetReader() Reader
	// Returns error that means skip item
	GetSkipItem() error
	// Returns error that means skip all
	GetSkipAll() error
	// Scans a file and calls a callback on each match.
	// Callback returns an error if occured. Error could be either SkipItem, or SkipAll, or any other error
	ScanFile(fileEntry DirEntry, searchRegexp *regexp.Regexp, callback func(SearchResult) error) error
	// Scans directories starting at the specified root path and calls a callback on each found entry.
	// Callback returns an error if occured. Error could be either SkipItem, or SkipAll, or any other error
	ScanDirs(rootPath string, depth int, callback func(DirEntry) error) error
}

// Handles search results
type Sink interface {
	// Handles search result
	HandleResult(result SearchResult)
}

// Performs search
type Searcher interface {
	// Returns scanner
	GetScanner() Scanner
	// Returns filter
	GetFilter() Filter
	// Returns sink
	GetSink() Sink
	// Starts search
	Search(ctx context.Context, rootPath string, searchRegexp *regexp.Regexp)
}