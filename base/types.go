package base

import (
	"context"
	"io"
	"regexp"
)

// Dir entry
type DirEntry struct {
	Path     string      // path to entry
	Depth    int         // recursion depth
	IsDir    bool        // whether the entry describes a directory
	Size     int64       // size of a file in bytes
}

// Represents a single match
type SearchResult struct {
	Path       string // path to file
	LineNumber int    // line number 1-based
	StartIndex int    // start index of a match 0-based
	EndIndex   int    // end index (exclusive) of a match 0-based
	Line       string // full line that has a match
}

// Search options
type SearchOptions struct {
	MaxSize       int64          // max size of file to scan in bytes
	MaxLength     int            // max length of a line to scan
	Include       *regexp.Regexp // include files that have matching path
	Exclude       *regexp.Regexp // exclude files that have matching path
	MatchCase     bool           // case-sensitivity
	Concurrency   int            // number of goroutines to spawn
	BufferSize    int            // size of buffers of channels
	MaxDepth      int            // max recursion depth
}

type Reader interface {
	OpenFile(fileEntry DirEntry) (interface { io.Reader; io.Closer }, error)
	ReadDir(dirEntry DirEntry) ([]DirEntry, error)
	ReadRootEntry(name string) (DirEntry, error)
}

type Scanner interface {
	ScanFile(fileEntry DirEntry, searchRegexp *regexp.Regexp, options SearchOptions, callback func(SearchResult) error) error
	ScanDir(rootPath string, options SearchOptions, callback func(DirEntry) error) error
}

type Sink interface {
	HandleResult(result SearchResult)
}

type Searcher interface {
	Search(rootPath string, searchRegexp *regexp.Regexp, options SearchOptions, ctx context.Context)
}