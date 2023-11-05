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
	defer finalizeProfile()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	reader := reader.NewFileSystemReader()
	var filterIns base.Filter
	if options.noSkip {
		filterIns = filter.NewNoopFilter()
	} else {
		filterIns = filter.NewConfigurableFilter(
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
	scanner := scanner.NewLineScanner(reader)
	sink := sink.NewWriterSink(os.Stdout)
	var searcherIns base.Searcher
	if options.concurrency == 0 {
		searcherIns = searcher.NewSerialSearcher(scanner, filterIns, sink, log.Default())
	} else {
		searcherIns = searcher.NewConcurrentSearcher(scanner, filterIns, sink, log.Default(), options.concurrency, options.bufferSize)
	}
	searcherIns.Search(ctx, searchDir, searchRegexp)
}