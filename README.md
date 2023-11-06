# mgrep

Find in files CLI tool writen in Go

## Usage

```
mgrep [OPTIONS] SEARCH [PATH]

SEARCH: regexp that will be tested on each line of scanned files

PATH: path to start scanning files

OPTIONS:

  -buf-size int
        Size of the buffers (default 1024)
  -concurr int
        How many concurrently running scanners to spawn (default 16)
  -exclude string
        Regexp of paths to exclude
  -include string
        Regexp of paths to include
  -match-case
        Match case
  -max-depth int
        Max recursion depth (default 100)
  -max-length int
        Max line length (default 1024)
  -max-size int
        Max file size in bytes (default 1048576)
  -no-subdirs
        Do not scan subdirectories. Same as max-depth=0
  -no-skip
        Do not skip anything
  -prof
        Run profiling. Set to cpu, heap, block, mutex or trace
```

## Build

Install go on your system. Run in command line from project root:

`go mod tidy`

`go build -v ./cmd/mgrep`

## Tests

To run tests execute this from project root:

`go test ./... -cover`

## Benchmarks

To run benchmarks execute this from project root:

`go test -bench=. -benchmem ./internal/searcher`