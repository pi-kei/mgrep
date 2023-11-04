package searcher

import (
	"context"
	"fmt"
	"regexp"
	"sync"

	"github.com/pi-kei/mgrep/internal/base"
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
	type pathAndDepth struct {
		path string
		depth int
	}
	pathsChannel := make(chan pathAndDepth, c.bufferSize)
	filesChannel := make(chan base.DirEntry, c.bufferSize)
	resultsChannel := make(chan base.SearchResult, c.bufferSize)

	var pathsWG sync.WaitGroup
	var dirsWG sync.WaitGroup
	dirsWG.Add(c.concurrency)
	for i := 0; i < c.concurrency; i++ {
		go func(index int) {
			defer dirsWG.Done()
			for {
				select {
				case newRootPath, ok := <-pathsChannel:
					if !ok {
						return
					}
					err := c.GetScanner().ScanDirs(newRootPath.path, newRootPath.depth, func(entry base.DirEntry) error {
						if entry.IsDir {
							if c.GetFilter().SkipDirEntry(entry) {
								return c.GetScanner().GetSkipItem()
							}
							if entry.Path == newRootPath.path {
								return nil
							}
							pathsWG.Add(1)
							select {
							case pathsChannel <- pathAndDepth{entry.Path, entry.Depth}:
								return c.GetScanner().GetSkipItem()
							case <-ctx.Done():
								pathsWG.Add(-1)
								return c.GetScanner().GetSkipAll()
							default:
								pathsWG.Add(-1)
								return nil
							}
						}
						
						if c.GetFilter().SkipFileEntry(entry) {
							return c.GetScanner().GetSkipItem()
						}
						select {
						case filesChannel <- entry:
							return nil
						case <-ctx.Done():
							return c.GetScanner().GetSkipAll()
						}
					})
					if err != nil {
						fmt.Println("Error scanning dir", err)
					}
					pathsWG.Done()
				case <-ctx.Done():
					return
				}
			}
		}(i)
	}
	go func() {
		defer close(filesChannel)
		dirsWG.Wait()
	}()

	var filesWG sync.WaitGroup
	filesWG.Add(c.concurrency)
	for i := 0; i < c.concurrency; i++ {
		go func() {
			defer filesWG.Done()
			for {
				select {
				case fileEntry, ok := <-filesChannel:
					if !ok {
						return
					}
					err := c.GetScanner().ScanFile(fileEntry, searchRegexp, func(sr base.SearchResult) error {
						if c.GetFilter().SkipSearchResult(sr) {
							return c.GetScanner().GetSkipItem()
						}
						select {
						case resultsChannel <- sr:
							return nil
						case <-ctx.Done():
							return c.GetScanner().GetSkipAll()
						}
					})
					if err != nil {
						fmt.Println("Error scanning file", err)
					}
				case <-ctx.Done():
					return
				}
			}
		}()
	}
	go func() {
		defer close(resultsChannel)
		filesWG.Wait()
	}()

	pathsWG.Add(1)
	pathsChannel <- pathAndDepth{rootPath, 0}
	go func() {
		defer close(pathsChannel)
		pathsWG.Wait()
	}()

	for result := range resultsChannel {
		c.GetSink().HandleResult(result)
	}
}