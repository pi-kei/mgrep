package searcher

import (
	"context"
	"fmt"
	"regexp"

	"github.com/pi-kei/mgrep/base"
	"github.com/pi-kei/mgrep/concurrency"
)

type Concurrent struct {
	scanner base.Scanner
	sink base.Sink
	concurrency   int            // number of goroutines to spawn
	bufferSize    int            // size of buffers of channels
}

func NewConcurrentSearcher(scanner base.Scanner, sink base.Sink, concurrency int, bufferSize int) base.Searcher {
	return &Concurrent{scanner, sink, concurrency, bufferSize}
}

func (c *Concurrent) GetScanner() base.Scanner {
	return c.scanner
}

func (c *Concurrent) GetSink() base.Sink {
	return c.sink
}

func (c *Concurrent) Search(rootPath string, searchRegexp *regexp.Regexp, ctx context.Context) {
	filesChannel := concurrency.Generator(func(handleDirEntry func(base.DirEntry) error) {
		err := c.GetScanner().ScanDirs(rootPath, handleDirEntry)
		if err != nil {
			fmt.Println("Error scanning dir", err)
		}
	}, c.bufferSize, ctx)

	filesChannels := concurrency.FanOut(filesChannel, c.concurrency, c.bufferSize, ctx)

	resultsChannels := concurrency.PipelineMulti(filesChannels, func(fileEntry base.DirEntry, handleSearchResult func(base.SearchResult) error) {
		err := c.GetScanner().ScanFile(fileEntry, searchRegexp, handleSearchResult)
		if err != nil {
			fmt.Println("Error scanning file", err)
		}
	}, c.bufferSize, ctx)

	resultsChannel := concurrency.FanIn(resultsChannels, c.bufferSize, ctx)

	for result := range resultsChannel {
		c.GetSink().HandleResult(result)
	}
}