package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"runtime"
	"syscall"

	"github.com/pi-kei/mgrep/base"
	"github.com/pi-kei/mgrep/reader"
	"github.com/pi-kei/mgrep/scanner"
	"github.com/pi-kei/mgrep/searcher"
	"github.com/pi-kei/mgrep/sink"
	"github.com/pi-kei/mgrep/skipper"
)

func parseArguments() (searchDir string, searchRegexp *regexp.Regexp, options base.SearchOptions) {
	maxSizeFlag := flag.Int64("max-size", 1024 * 1024, "Max file size in bytes")
	maxLengthFlag := flag.Int("max-length", 1024, "Max line length")
	includeFlag := flag.String("include", "", "Regexp of paths to include")
	excludeFlag := flag.String("exclude", "", "Regexp of paths to exclude")
	matchCaseFlag := flag.Bool("match-case", false, "Match case")
	noSubdirsFlag := flag.Bool("no-subdirs", false, "Do not scan subdirectories. Same as max-depth=0")
	concurrFlag := flag.Int("concurr", runtime.NumCPU(), "How many concurrently running scanners to spawn")
	bufferSizeFlag := flag.Int("buf-size", 1024, "Size of the buffers")
	maxDepthFlag := flag.Int("max-depth", 100, "Max recursion depth")

	flag.Parse()

	if flag.NArg() < 1 || flag.NArg() > 2 {
		fmt.Println("Expecting search string and optionally search dir arguments")
		os.Exit(1)
	}

	searchDir = "."
	if flag.NArg() == 2 {
		searchDir = flag.Arg(1)
	}

	options = base.SearchOptions{
		MaxSize: *maxSizeFlag,
		MaxLength: *maxLengthFlag,
		Include: nil,
		Exclude: nil,
		MatchCase: *matchCaseFlag,
		Concurrency: *concurrFlag,
		BufferSize: *bufferSizeFlag,
		MaxDepth: *maxDepthFlag,
	}

	if len(*includeFlag) > 0 {
		include, err := regexp.Compile(*includeFlag)
		if err != nil {
			fmt.Println("Invalid include", err)
			os.Exit(1)
		}
		options.Include = include
	}

	if len(*excludeFlag) > 0 {
		exclude, err := regexp.Compile(*excludeFlag)
		if err != nil {
			fmt.Println("Invalid exclude", err)
			os.Exit(1)
		}
		options.Exclude = exclude
	}

	searchPattern := flag.Arg(0)
	if !options.MatchCase {
		searchPattern = "(?i)" + searchPattern
	}
	
	searchRegexp, err := regexp.Compile(searchPattern)
	if err != nil {
		fmt.Println("Invalid search pattern", err)
		os.Exit(1)
	}

	if options.MaxSize < 1 {
		options.MaxSize = 1
	}

	if options.MaxLength < 1 {
		options.MaxLength = 1
	}

	if options.Concurrency < 1 {
		options.Concurrency = 1
	}

	if options.BufferSize < 0 {
		options.BufferSize = 0
	}

	if options.MaxDepth < 0 || *noSubdirsFlag {
		options.MaxDepth = 0
	}

	return searchDir, searchRegexp, options
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	searchDir, searchRegexp, options := parseArguments()
	reader := reader.NewFileSystemReader()
	skipper := skipper.NewConfigurableSkipper()
	scanner := scanner.NewLineScanner(reader)
	sink := sink.NewStdoutSink()
	searcher := searcher.NewConcurrentSearcher(scanner, skipper, sink)
	searcher.Search(searchDir, searchRegexp, options, ctx)
}