package sink

import (
	"log"

	"github.com/pi-kei/mgrep/internal/base"
)

type Logger struct {
	logger *log.Logger
	format string
	getValues func(result base.SearchResult) []any
}

// Sink that writes formatted strings using specified logger.
// Uses default format.
// Thread-safe.
func NewLoggerSink(logger *log.Logger) base.Sink {
	return &Logger{logger, DefaultFormat, DefaultGetValues}
}

// Sink that writes formatted strings using specified logger.
// Uses specified format.
// Thread-safe.
func NewCustomLoggerSink(logger *log.Logger, format string, getValues func(result base.SearchResult) []any) base.Sink {
	return &Logger{logger, format, getValues}
}

func (l *Logger) HandleResult(result base.SearchResult) {
	l.logger.Printf(l.format, l.getValues(result)...)
}