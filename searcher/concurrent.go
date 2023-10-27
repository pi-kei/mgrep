package searcher

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"sync"

	"github.com/pi-kei/mgrep/base"
)

type Concurrent struct {
	scanner base.Scanner
	skipper base.Skipper
	sink base.Sink
}

func NewConcurrentSearcher(scanner base.Scanner, skipper base.Skipper, sink base.Sink) base.Searcher {
	return &Concurrent{scanner, skipper, sink}
}

func (c *Concurrent) Search(rootPath string, searchRegexp *regexp.Regexp, options base.SearchOptions, ctx context.Context) {
	filesChannel := make(chan base.DirEntry, options.BufferSize)

	go func() {
		defer close(filesChannel)
		err := c.scanner.ScanDir(rootPath, options, func(fileEntry base.DirEntry) error {
			if fileEntry.IsDir {
				if c.skipper.SkipDirEntry(fileEntry, options) {
					return base.SkipDirEntry
				}
				return nil
			}
			if c.skipper.SkipFileEntry(fileEntry, options) {
				return base.SkipFileEntry
			}
			select {
			case filesChannel <- fileEntry:
				return nil
			case <-ctx.Done():
				return base.SkipAll
			}
		})
		if err != nil && err != filepath.SkipAll {
			fmt.Println("Error scanning dir", err)
		}
	}()

	resultsChannel := make(chan base.SearchResult, options.BufferSize)
	var resultsWG sync.WaitGroup

	for i := 0; i < options.Concurrency; i++ {
		resultsWG.Add(1)
		go func() {
			defer resultsWG.Done()
			for fileEntry := range filesChannel {
				err := c.scanner.ScanFile(fileEntry, searchRegexp, options, func(result base.SearchResult) error {
					if c.skipper.SkipSearchResult(result, options) {
						return base.SkipSearchResult
					}
					select {
					case resultsChannel <- result:
						return nil
					case <-ctx.Done():
						return base.SkipAll
					}
				})
				if err != nil {
					fmt.Println("Error scanning file", err)
				}
			}
		}()
	}

	go func() {
		defer close(resultsChannel)
		resultsWG.Wait()
	}()

	for result := range resultsChannel {
		c.sink.HandleResult(result)
	}
}