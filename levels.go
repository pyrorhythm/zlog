package zlog

import (
	"log/slog"
	"os"
)

const (
	LevelTrace = slog.Level(-8 + 4*iota)
	LevelDebug
	LevelInfo
	LevelWarn
	LevelError
	LevelPanic
	LevelFatal
)

// NewLevels is a function that maps custom level values to their names.
func NewLevels(a slog.Attr) slog.Attr {
	if a.Key == slog.LevelKey {
		if logLevel, ok := a.Value.Any().(slog.Level); ok {
			a.Value = slog.StringValue(levelName(logLevel))
		}
	}
	return a
}

func levelName(level slog.Level) string {
	switch level {
	case LevelTrace:
		return "TRACE"
	case LevelPanic:
		return "PANIC"
	case LevelFatal:
		return "FATAL"
	default:
		return level.String()
	}
}

func executeLevelSpecificActions(r slog.Record) {
	switch r.Level {
	case LevelPanic:
		panic(r.Message)
	case LevelFatal:
		os.Exit(1)
	}
}
