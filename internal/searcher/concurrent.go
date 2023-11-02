package searcher

import (
	"context"
	"errors"
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

func (c *Concurrent) Search(ctx context.Context, rootPath string, searchRegexp *regexp.Regexp) {
	filesChannels := concurrency.ProcRecursively(ctx, rootPath, func(newRootPath string, send func(string) (int, error), handleDirEntry func(base.DirEntry) error) {
		err := c.GetScanner().ScanDirs(newRootPath, func(entry base.DirEntry) error {
			if entry.IsDir {
				if c.GetFilter().SkipDirEntry(entry) {
					return c.GetScanner().GetSkipItem()
				}
				if entry.Path == newRootPath {
					return nil
				}
				chosen, err := send(entry.Path)
				if err != nil {
					return c.GetScanner().GetSkipAll()
				}
				if chosen >= 0 {
					return c.GetScanner().GetSkipItem()
				}
				return nil
			}
			
			if c.GetFilter().SkipFileEntry(entry) {
				return c.GetScanner().GetSkipItem()
			}
			err := handleDirEntry(entry)
			if ctx.Err() != nil && errors.Is(err, ctx.Err()) {
				return c.GetScanner().GetSkipAll()
			}
			return err
		})
		if err != nil {
			fmt.Println("Error scanning dir", err)
		}
	}, c.concurrency, c.bufferSize)

	resultsChannels := concurrency.PipelineMulti(ctx, filesChannels, func(fileEntry base.DirEntry, handleSearchResult func(base.SearchResult) error) {
		err := c.GetScanner().ScanFile(fileEntry, searchRegexp, func(sr base.SearchResult) error {
			if c.GetFilter().SkipSearchResult(sr) {
				return c.GetScanner().GetSkipItem()
			}
			err := handleSearchResult(sr)
			if ctx.Err() != nil && errors.Is(err, ctx.Err()) {
				return c.GetScanner().GetSkipAll()
			}
			return err
		})
		if err != nil {
			fmt.Println("Error scanning file", err)
		}
	}, c.bufferSize)

	resultsChannel := concurrency.FanIn(ctx, resultsChannels, c.bufferSize)

	for result := range resultsChannel {
		c.GetSink().HandleResult(result)
	}
}