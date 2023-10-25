package main

import (
	"os"
	"regexp"
)

// Dir entry
type DirEntry struct {
	path     string      // path to entry
	depth    int         // recursion depth
	dirEntry os.DirEntry // entry struct
}

// Represents a single match
type SearchResult struct {
	path       string // path to file
	lineNumber int    // line number 1-based
	startIndex int    // start index of a match 0-based
	endIndex   int    // end index (exclusive) of a match 0-based
	line       string // full line that has a match
}

// Search options
type SearchOptions struct {
	maxSize       int64          // max size of file to scan in bytes
	maxLength     int            // max length of a line to scan
	include       *regexp.Regexp // include files that have matching path
	exclude       *regexp.Regexp // exclude files that have matching path
	matchCase     bool           // case-sensitivity
	concurrency   int            // number of goroutines to spawn
	bufferSize    int            // size of buffers of channels
	maxDepth      int            // max recursion depth
}