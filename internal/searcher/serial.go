package searcher

import (
	"context"
	"log"
	"regexp"

	"github.com/pi-kei/mgrep/internal/base"
)

type Serial struct {
	scanner base.Scanner
	filter  base.Filter
	sink    base.Sink
	logger  *log.Logger
}

func NewSerial(scanner base.Scanner, filter base.Filter, sink base.Sink, logger *log.Logger) base.Searcher {
	return &Serial{scanner, filter, sink, logger}
}

func (s *Serial) Search(ctx context.Context, rootPath string, searchRegexp *regexp.Regexp) {
	done := make(chan struct{})

	go func() {
		err := s.scanner.ScanDirs(rootPath, 0, func(entry base.DirEntry) error {
			select {
			case <-ctx.Done():
				return base.ErrSkipAll
			default:
			}

			if entry.IsDir {
				if s.filter.SkipDirEntry(entry) {
					return base.ErrSkipItem
				}
				return nil
			}

			if s.filter.SkipFileEntry(entry) {
				return base.ErrSkipItem
			}
			err := s.scanner.ScanFile(entry, searchRegexp, func(result base.SearchResult) error {
				select {
				case <-ctx.Done():
					return base.ErrSkipAll
				default:
				}

				if s.filter.SkipSearchResult(result) {
					return base.ErrSkipItem
				}
				s.sink.HandleResult(result)
				return nil
			})
			if err != nil {
				s.logger.Println("Error scanning file", err)
			}
			return nil
		})
		if err != nil {
			s.logger.Println("Error scanning dir", err)
		}
		done <- struct{}{}
	}()

	<-done
}
