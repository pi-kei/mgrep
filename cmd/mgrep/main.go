package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"unicode/utf8"

	"github.com/pi-kei/mgrep/internal/base"
	"github.com/pi-kei/mgrep/internal/filter"
	"github.com/pi-kei/mgrep/internal/reader"
	"github.com/pi-kei/mgrep/internal/scanner"
	"github.com/pi-kei/mgrep/internal/searcher"
	"github.com/pi-kei/mgrep/internal/sink"
)

func main() {
	searchDir, searchRegexp, options := parseArguments()

	finalizeProfile, err := getProfile(options.profile)
	if err != nil {
		log.Println(err)
		return
	}
	if finalizeProfile != nil {
		defer finalizeProfile()
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	searcherIns := buildSearcher(options)
	searcherIns.Search(ctx, searchDir, searchRegexp)
}

func buildSearcher(options searchOptions) base.Searcher {
	reader := reader.NewFileSystem()
	var filterIns base.Filter
	if options.noSkip {
		filterIns = filter.NewNoop()
	} else {
		filterIns = filter.NewConfigurable(
			func(dirEntry base.DirEntry) bool {
				return dirEntry.Depth > options.maxDepth
			},
			func(fileEntry base.DirEntry) bool {
				if fileEntry.Size == 0 {
					return true
				}
				if fileEntry.Size > options.maxSize {
					return true
				}
				if options.include != nil && !options.include.MatchString(fileEntry.Path) {
					return true
				}
				if options.exclude != nil && options.exclude.MatchString(fileEntry.Path) {
					return true
				}
				return false
			},
			func(searchResult base.SearchResult) bool {
				return utf8.RuneCountInString(searchResult.Line) > options.maxLength
			},
		)
	}
	scanner := scanner.NewLine(reader)
	sink := sink.NewWriter(os.Stdout)
	var searcherIns base.Searcher
	if options.concurrency == 0 {
		searcherIns = searcher.NewSerial(scanner, filterIns, sink, log.Default())
	} else {
		searcherIns = searcher.NewConcurrent(scanner, filterIns, sink, log.Default(), options.concurrency, options.bufferSize)
	}
	return searcherIns
}
