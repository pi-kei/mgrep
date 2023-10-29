package searcher

import (
	"context"
	"fmt"
	"regexp"

	"github.com/pi-kei/mgrep/internal/base"
)

type Linear struct {
	scanner base.Scanner
	filter  base.Filter
	sink    base.Sink
}

func NewLinearSearcher(scanner base.Scanner, filter base.Filter, sink base.Sink) base.Searcher {
	return &Linear{scanner, filter, sink}
}

func (l *Linear) GetScanner() base.Scanner {
	return l.scanner
}

func (l *Linear) GetFilter() base.Filter {
	return l.filter
}

func (l *Linear) GetSink() base.Sink {
	return l.sink
}

func (l *Linear) Search(rootPath string, searchRegexp *regexp.Regexp, ctx context.Context) {
	done := make(chan struct{})

	go func() {
		err := l.GetScanner().ScanDirs(rootPath, func(entry base.DirEntry) error {
			select {
			case <-ctx.Done():
				return l.GetScanner().GetSkipAll()
			default:
			}
			
			if entry.IsDir {
				if l.GetFilter().SkipDirEntry(entry) {
					return l.GetScanner().GetSkipItem()
				}
				return nil
			}
			
			if l.GetFilter().SkipFileEntry(entry) {
				return l.GetScanner().GetSkipItem()
			}
			err := l.GetScanner().ScanFile(entry, searchRegexp, func(result base.SearchResult) error {
				select {
				case <-ctx.Done():
					return l.GetScanner().GetSkipAll()
				default:
				}
				
				if l.GetFilter().SkipSearchResult(result) {
					return l.GetScanner().GetSkipItem()
				}
				l.GetSink().HandleResult(result)
				return nil
			})
			if err != nil {
				fmt.Println("Error scanning file", err)
			}
			return nil
		})
		if err != nil {
			fmt.Println("Error scanning dir", err)
		}
		done <- struct{}{}
	}()
	
	<-done
}