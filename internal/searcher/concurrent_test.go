package searcher

import (
	"context"
	"regexp"
	"runtime"
	"testing"
	"time"

	"github.com/pi-kei/mgrep/internal/filter"
	"github.com/pi-kei/mgrep/internal/reader"
	"github.com/pi-kei/mgrep/internal/scanner"
	"github.com/pi-kei/mgrep/internal/sink"
)

func BenchmarkConcurrentSearcher(b *testing.B) {
	b.Run("1", func(b *testing.B) {
		benchmarkConcurrentSearcher(b, 10, 10, 3, 5, 5)
	})
	b.Run("2", func(b *testing.B) {
		benchmarkConcurrentSearcher(b, 18, 50, 4, 7, 7)
	})
	b.Run("3", func(b *testing.B) {
		benchmarkConcurrentSearcher(b, 13, 75, 5, 8, 8)
	})
	b.Run("4", func(b *testing.B) {
		benchmarkConcurrentSearcher(b, 15, 100, 6, 10, 10)
	})
	b.Run("5", func(b *testing.B) {
		benchmarkConcurrentSearcher(b, 92, 100, 7, 9, 9)
	})
}

func benchmarkConcurrentSearcher(b *testing.B, seed int64, maxLines, maxDepth, maxDirs, maxFiles int) {
	entries, rootName, _ := reader.NewEntriesGen(seed, maxLines, maxDepth, maxDirs, maxFiles, time.Now().UTC(), 48).Generate()
	b.Logf("Entries generated: %v", len(entries))
	reader := reader.NewMockReader(entries)
	scanner := scanner.NewLineScanner(reader)
	filter := filter.NewNoopFilter()
	sink := sink.NewNoopSink()
	searcher := NewConcurrentSearcher(scanner, filter, sink, runtime.NumCPU(), 1024)
	ctx := context.Background()
	re := regexp.MustCompile("and")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		searcher.Search(ctx, rootName, re)
	}
}