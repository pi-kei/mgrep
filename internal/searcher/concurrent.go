package searcher

import (
	"context"
	"fmt"
	"regexp"

	"github.com/pi-kei/mgrep/internal/base"
	"github.com/pi-kei/mgrep/pkg/concurrency"
)

type Concurrent struct {
	scanner base.Scanner
	filter base.Filter
	sink base.Sink
	concurrency   int            // number of goroutines to spawn
	bufferSize    int            // size of buffers of channels
}

func NewConcurrentSearcher(scanner base.Scanner, filter base.Filter, sink base.Sink, concurrency int, bufferSize int) base.Searcher {
	return &Concurrent{scanner, filter, sink, concurrency, bufferSize}
}

func (c *Concurrent) GetScanner() base.Scanner {
	return c.scanner
}

func (c *Concurrent) GetFilter() base.Filter {
	return c.filter
}

func (c *Concurrent) GetSink() base.Sink {
	return c.sink
}

func (c *Concurrent) Search(rootPath string, searchRegexp *regexp.Regexp, ctx context.Context) {
	filesChannels := concurrency.ProcRecursively(rootPath, func(newRootPath string, send func(string) (bool, error), handleDirEntry func(base.DirEntry) error) {
		err := c.GetScanner().ScanDirs(newRootPath, func(entry base.DirEntry) (bool, error) {
			if entry.IsDir {
				if skipDir := c.GetFilter().SkipDirEntry(entry); skipDir {
					return true, nil
				}
				if entry.Path == newRootPath {
					return false, nil
				}
				sent, err := send(entry.Path)
				if err != nil {
					return false, err
				}
				if sent {
					return true, nil
				}
				return false, nil
			}
			if skipFile := c.GetFilter().SkipFileEntry(entry); skipFile {
				return true, nil
			}
			err := handleDirEntry(entry)
			return false, err
		})
		if err != nil && ctx.Err() == nil {
			fmt.Println("Error scanning dir", err)
		}
	}, c.concurrency, c.bufferSize, ctx)

	resultsChannels := concurrency.PipelineMulti(filesChannels, func(fileEntry base.DirEntry, handleSearchResult func(base.SearchResult) error) {
		err := c.GetScanner().ScanFile(fileEntry, searchRegexp, func(sr base.SearchResult) (bool, error) {
			if skipResult := c.GetFilter().SkipSearchResult(sr); skipResult {
				return true, nil
			}
			err := handleSearchResult(sr)
			return false, err
		})
		if err != nil && ctx.Err() == nil {
			fmt.Println("Error scanning file", err)
		}
	}, c.bufferSize, ctx)

	resultsChannel := concurrency.FanIn(resultsChannels, c.bufferSize, ctx)

	for result := range resultsChannel {
		c.GetSink().HandleResult(result)
	}
}