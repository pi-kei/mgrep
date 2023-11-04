package searcher

import (
	"context"
	"fmt"
	"regexp"

	"github.com/pi-kei/mgrep/internal/base"
)

type Serial struct {
	scanner base.Scanner
	filter  base.Filter
	sink    base.Sink
}

func NewSerialSearcher(scanner base.Scanner, filter base.Filter, sink base.Sink) base.Searcher {
	return &Serial{scanner, filter, sink}
}

func (s *Serial) GetScanner() base.Scanner {
	return s.scanner
}

func (s *Serial) GetFilter() base.Filter {
	return s.filter
}

func (s *Serial) GetSink() base.Sink {
	return s.sink
}

func (s *Serial) Search(ctx context.Context, rootPath string, searchRegexp *regexp.Regexp) {
	done := make(chan struct{})

	go func() {
		err := s.GetScanner().ScanDirs(rootPath, 0, func(entry base.DirEntry) error {
			select {
			case <-ctx.Done():
				return s.GetScanner().GetSkipAll()
			default:
			}
			
			if entry.IsDir {
				if s.GetFilter().SkipDirEntry(entry) {
					return s.GetScanner().GetSkipItem()
				}
				return nil
			}
			
			if s.GetFilter().SkipFileEntry(entry) {
				return s.GetScanner().GetSkipItem()
			}
			err := s.GetScanner().ScanFile(entry, searchRegexp, func(result base.SearchResult) error {
				select {
				case <-ctx.Done():
					return s.GetScanner().GetSkipAll()
				default:
				}
				
				if s.GetFilter().SkipSearchResult(result) {
					return s.GetScanner().GetSkipItem()
				}
				s.GetSink().HandleResult(result)
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