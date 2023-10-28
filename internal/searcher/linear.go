package searcher

import (
	"context"
	"fmt"
	"regexp"

	"github.com/pi-kei/mgrep/internal/base"
)

type Linear struct {
	scanner base.Scanner
	sink    base.Sink
}

func NewLinearSearcher(scanner base.Scanner, sink base.Sink) base.Searcher {
	return &Linear{scanner, sink}
}

func (l *Linear) GetScanner() base.Scanner {
	return l.scanner
}

func (l *Linear) GetSink() base.Sink {
	return l.sink
}

func (l *Linear) Search(rootPath string, searchRegexp *regexp.Regexp, ctx context.Context) {
	err := l.GetScanner().ScanDirs(rootPath, func(fileEntry base.DirEntry) error {
		if fileEntry.IsDir {
			return nil
		}
		err := l.GetScanner().ScanFile(fileEntry, searchRegexp, func(result base.SearchResult) error {
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
}