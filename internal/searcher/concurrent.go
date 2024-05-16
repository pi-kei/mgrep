package searcher

import (
	"context"
	"log"
	"regexp"
	"sync"

	"github.com/pi-kei/mgrep/internal/base"
)

type Concurrent struct {
	scanner     base.Scanner
	filter      base.Filter
	sink        base.Sink
	logger      *log.Logger
	concurrency int // number of goroutines to spawn
	bufferSize  int // size of buffers of channels
}

func NewConcurrent(scanner base.Scanner, filter base.Filter, sink base.Sink, logger *log.Logger, concurrency int, bufferSize int) base.Searcher {
	return &Concurrent{scanner, filter, sink, logger, concurrency, bufferSize}
}

func (c *Concurrent) Search(ctx context.Context, rootPath string, searchRegexp *regexp.Regexp) {
	type pathAndDepth struct {
		path  string
		depth int
	}
	pathsChannel := make(chan pathAndDepth, c.bufferSize)
	filesChannel := make(chan base.DirEntry, c.bufferSize)
	resultsChannel := make(chan base.SearchResult, c.bufferSize)

	dirsConcurr := 1
	filesConcurr := 1
	if c.concurrency > 2 {
		dirsConcurr = c.concurrency / 2
		filesConcurr = c.concurrency - dirsConcurr
	}

	var pathsWG sync.WaitGroup
	var dirsWG sync.WaitGroup
	dirsWG.Add(dirsConcurr)
	for i := 0; i < dirsConcurr; i++ {
		go func(index int) {
			defer dirsWG.Done()
			for {
				select {
				case newRootPath, ok := <-pathsChannel:
					if !ok {
						return
					}
					err := c.scanner.ScanDirs(newRootPath.path, newRootPath.depth, func(entry base.DirEntry) error {
						if entry.IsDir {
							if c.filter.SkipDirEntry(entry) {
								return base.ErrSkipItem
							}
							if entry.Path == newRootPath.path {
								return nil
							}
							pathsWG.Add(1)
							select {
							case pathsChannel <- pathAndDepth{entry.Path, entry.Depth}:
								return base.ErrSkipItem
							case <-ctx.Done():
								pathsWG.Add(-1)
								return base.ErrSkipAll
							default:
								pathsWG.Add(-1)
								return nil
							}
						}

						if c.filter.SkipFileEntry(entry) {
							return base.ErrSkipItem
						}
						select {
						case filesChannel <- entry:
							return nil
						case <-ctx.Done():
							return base.ErrSkipAll
						}
					})
					if err != nil {
						c.logger.Println("Error scanning dir", err)
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
	filesWG.Add(filesConcurr)
	for i := 0; i < filesConcurr; i++ {
		go func() {
			defer filesWG.Done()
			for {
				select {
				case fileEntry, ok := <-filesChannel:
					if !ok {
						return
					}
					err := c.scanner.ScanFile(fileEntry, searchRegexp, func(sr base.SearchResult) error {
						if c.filter.SkipSearchResult(sr) {
							return base.ErrSkipItem
						}
						select {
						case resultsChannel <- sr:
							return nil
						case <-ctx.Done():
							return base.ErrSkipAll
						}
					})
					if err != nil {
						c.logger.Println("Error scanning file", err)
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
		c.sink.HandleResult(result)
	}
}
