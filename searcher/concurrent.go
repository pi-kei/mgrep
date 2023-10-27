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
	sink base.Sink
	concurrency   int            // number of goroutines to spawn
	bufferSize    int            // size of buffers of channels
}

func NewConcurrentSearcher(scanner base.Scanner, sink base.Sink, concurrency int, bufferSize int) base.Searcher {
	return &Concurrent{scanner, sink, concurrency, bufferSize}
}

func (c *Concurrent) Search(rootPath string, searchRegexp *regexp.Regexp, ctx context.Context) {
	filesChannel := make(chan base.DirEntry, c.bufferSize)
	resultsChannel := make(chan base.SearchResult, c.bufferSize)

	var resultsWG sync.WaitGroup

	go func() {
		defer close(filesChannel)
		err := c.scanner.ScanDirs(rootPath, func(fileEntry base.DirEntry) error {
			select {
			case filesChannel <- fileEntry:
				return nil
			case <-ctx.Done():
				return filepath.SkipAll
			}
		})
		if err != nil {
			fmt.Println("Error scanning dir", err)
		}
	}()

	for i := 0; i < c.concurrency; i++ {
		resultsWG.Add(1)
		go func() {
			defer resultsWG.Done()
			for fileEntry := range filesChannel {
				err := c.scanner.ScanFile(fileEntry, searchRegexp, func(result base.SearchResult) error {
					select {
					case resultsChannel <- result:
						return nil
					case <-ctx.Done():
						return filepath.SkipAll
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