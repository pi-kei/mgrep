package searcher

import (
	"context"
	"log"
	"regexp"
	"testing"
	"time"

	"github.com/pi-kei/mgrep/internal/filter"
	"github.com/pi-kei/mgrep/internal/reader"
	"github.com/pi-kei/mgrep/internal/scanner"
	"github.com/pi-kei/mgrep/internal/sink"
)

func BenchmarkSerialSearcher(b *testing.B) {
	b.Run("1", func(b *testing.B) {
		benchmarkSerialSearcher(b, 103, 5, 2, 4, 4)
	})
	b.Run("2", func(b *testing.B) {
		benchmarkSerialSearcher(b, 10, 10, 3, 5, 5)
	})
	b.Run("3", func(b *testing.B) {
		benchmarkSerialSearcher(b, 18, 50, 4, 7, 7)
	})
	b.Run("4", func(b *testing.B) {
		benchmarkSerialSearcher(b, 13, 75, 5, 8, 8)
	})
	b.Run("5", func(b *testing.B) {
		benchmarkSerialSearcher(b, 15, 100, 6, 9, 9)
	})
	b.Run("6", func(b *testing.B) {
		benchmarkSerialSearcher(b, 92, 150, 7, 10, 10)
	})
	b.Run("7", func(b *testing.B) {
		benchmarkSerialSearcher(b, 92, 200, 8, 11, 11)
	})
	b.Run("8", func(b *testing.B) {
		benchmarkSerialSearcher(b, 145, 0, 9, 12, 0)
	})
	b.Run("9", func(b *testing.B) {
		benchmarkSerialSearcher(b, 145, 250, 2, 1, 100000)
	})
}

func benchmarkSerialSearcher(b *testing.B, seed int64, maxLines, maxDepth, maxDirs, maxFiles int) {
	entries, rootName, contents := reader.NewEntriesGen(seed, maxLines, maxDepth, maxDirs, maxFiles, time.Now().UTC(), 48).Generate()
	b.Logf("Entries generated: %d (files: %d)", len(entries), len(contents))
	reader := reader.NewMockReader(entries)
	scanner := scanner.NewLine(reader)
	filter := filter.NewNoop()
	sink := sink.NewNoop()
	searcher := NewSerial(scanner, filter, sink, log.Default())
	ctx := context.Background()
	re := regexp.MustCompile("and")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		searcher.Search(ctx, rootName, re)
	}
}
