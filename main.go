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
)

func parseArguments() (searchDir string, searchRegexp *regexp.Regexp, options SearchOptions) {
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

	options = SearchOptions{
		maxSize: *maxSizeFlag,
		maxLength: *maxLengthFlag,
		include: nil,
		exclude: nil,
		matchCase: *matchCaseFlag,
		concurrency: *concurrFlag,
		bufferSize: *bufferSizeFlag,
		maxDepth: *maxDepthFlag,
	}

	if len(*includeFlag) > 0 {
		include, err := regexp.Compile(*includeFlag)
		if err != nil {
			fmt.Println("Invalid include", err)
			os.Exit(1)
		}
		options.include = include
	}

	if len(*excludeFlag) > 0 {
		exclude, err := regexp.Compile(*excludeFlag)
		if err != nil {
			fmt.Println("Invalid exclude", err)
			os.Exit(1)
		}
		options.exclude = exclude
	}

	searchPattern := flag.Arg(0)
	if !options.matchCase {
		searchPattern = "(?i)" + searchPattern
	}
	
	searchRegexp, err := regexp.Compile(searchPattern)
	if err != nil {
		fmt.Println("Invalid search pattern", err)
		os.Exit(1)
	}

	if options.maxSize < 1 {
		options.maxSize = 1
	}

	if options.maxLength < 1 {
		options.maxLength = 1
	}

	if options.concurrency < 1 {
		options.concurrency = 1
	}

	if options.bufferSize < 0 {
		options.bufferSize = 0
	}

	if options.maxDepth < 0 || *noSubdirsFlag {
		options.maxDepth = 0
	}

	return searchDir, searchRegexp, options
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	searchDir, searchRegexp, options := parseArguments()
	search(searchDir, searchRegexp, options, ctx)
}