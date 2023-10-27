package searcher

import (
	"context"
	"fmt"
	"regexp"

	"github.com/pi-kei/mgrep/base"
)

type Linear struct {
	scanner base.Scanner
	sink    base.Sink
}

func NewLinearSearcher(scanner base.Scanner, sink base.Sink) base.Searcher {
	return &Linear{scanner, sink}
}

func (l *Linear) Search(rootPath string, searchRegexp *regexp.Regexp, ctx context.Context) {
	err := l.scanner.ScanDirs(rootPath, func(fileEntry base.DirEntry) error {
		err := l.scanner.ScanFile(fileEntry, searchRegexp, func(result base.SearchResult) error {
			l.sink.HandleResult(result)
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