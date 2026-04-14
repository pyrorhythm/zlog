package zlog

import (
	"log/slog"
	"os"

	"github.com/fatih/color"
)

var colorMap = map[slog.Level]*color.Color{
	LevelTrace: color.RGB(67, 120, 198),
	LevelDebug: color.RGB(65, 185, 198),
	LevelInfo:  color.RGB(134, 180, 249),
	LevelWarn:  color.RGB(252, 172, 75),
	LevelError: color.RGB(252, 99, 75),
	LevelPanic: color.RGB(216, 52, 0),
	LevelFatal: color.RGB(73, 0, 0),
}

const (
	LevelTrace = slog.Level(-8 + 4*iota)
	LevelDebug
	LevelInfo
	LevelWarn
	LevelError
	LevelPanic
	LevelFatal
)

func NewLevels(_ []string, a slog.Attr) slog.Attr {
	if a.Key == slog.LevelKey {
		logLevel := a.Value.Any().(slog.Level)
		levelLabel, ok := newLevelsStringer(logLevel)
		if !ok {
			levelLabel = logLevel.String()
		}

		a.Value = slog.StringValue(levelLabel)
	}

	return a
}

func newLevelsStringer(lvlr slog.Leveler) (string, bool) {
	switch lvlr.Level() {
	case LevelTrace:
		return "TRACE", true
	case LevelPanic:
		return "PANIC", true
	case LevelFatal:
		return "FATAL", true
	default:
		return "", false
	}
}

func executeLevelSpecificActions(r slog.Record) {
	switch r.Level {
	case LevelPanic:
		panic(r.Message)
	case LevelFatal:
		os.Exit(1)
	default:
	}
}

func formatLevel(r slog.Record) string {
	level := r.Level.String()
	col, ok := colorMap[r.Level]
	if ok {
		level = col.Sprint(level)
	}
	return level
}
