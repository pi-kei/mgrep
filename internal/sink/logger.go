package sink

import (
	"log"

	"github.com/pi-kei/mgrep/internal/base"
)

type Logger struct {
	logger    *log.Logger
	format    string
	getValues func(result base.SearchResult) []any
}

type LoggerOption func(*Logger)

func WithLoggerFormat(format string) LoggerOption {
	return func(l *Logger) {
		l.format = format
	}
}

func WithLoggerGetValues(getValues func(result base.SearchResult) []any) LoggerOption {
	return func(l *Logger) {
		l.getValues = getValues
	}
}

// Sink that writes formatted strings using specified logger.
// Thread-safe.
func NewLogger(logger *log.Logger, options ...LoggerOption) base.Sink {
	sink := Logger{logger, DefaultFormat, DefaultGetValues}
	for _, option := range options {
		option(&sink)
	}
	return &sink
}

func (l *Logger) HandleResult(result base.SearchResult) {
	l.logger.Printf(l.format, l.getValues(result)...)
}
