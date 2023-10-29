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
		err := l.GetScanner().ScanDirs(rootPath, func(entry base.DirEntry) (bool, error) {
			select {
			case <-ctx.Done():
				return false, ctx.Err()
			default:
			}
			if entry.IsDir {
				return l.GetFilter().SkipDirEntry(entry), nil
			}
			if skipFile := l.GetFilter().SkipFileEntry(entry); skipFile {
				return true, nil
			}
			err := l.GetScanner().ScanFile(entry, searchRegexp, func(result base.SearchResult) (bool, error) {
				select {
				case <-ctx.Done():
					return false, ctx.Err()
				default:
				}
				if skipResult := l.GetFilter().SkipSearchResult(result); skipResult {
					return true, nil
				}
				l.GetSink().HandleResult(result)
				return false, nil
			})
			if err != nil && ctx.Err() == nil {
				fmt.Println("Error scanning file", err)
			}
			return false, nil
		})
		if err != nil && ctx.Err() == nil {
			fmt.Println("Error scanning dir", err)
		}
		done <- struct{}{}
	}()
	
	<-done
}