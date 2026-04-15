package zlog

import (
	"io"
	"log/slog"
	"os"
)

// Sink configures a single log output destination.
type Sink struct {
	Writer io.Writer
	Level  slog.Level
	Style  Style
}

// Options configures the logger.
type Options struct {
	// Middleware is applied globally to every record before routing to sinks.
	Middleware []Middleware
	// Sinks defines output destinations.
	Sinks []Sink
}

// New creates a *slog.Logger. Without arguments, outputs colored pretty-print to stdout.
func New(opts ...Options) *slog.Logger {
	var opt Options
	if len(opts) > 0 {
		opt = opts[0]
	}

	if len(opt.Middleware) == 0 {
		opt.Middleware = defaultMiddleware()
	}

	if len(opt.Sinks) == 0 {
		opt.Sinks = []Sink{{
			Writer: os.Stdout,
			Level:  LevelDebug,
			Style:  DefaultStyle,
		}}
	}

	handlers := make([]slog.Handler, len(opt.Sinks))
	for i, sink := range opt.Sinks {
		if sink.Writer == nil {
			sink.Writer = os.Stdout
		}
		handlers[i] = newHandler(sink.Writer, sink.Level, sink.Style, opt.Middleware)
	}

	if len(handlers) == 1 {
		return slog.New(handlers[0])
	}
	return slog.New(&multiHandler{handlers: handlers})
}

func defaultMiddleware() []Middleware {
	return []Middleware{
		ContextMiddleware,
		ReplaceAttributeMiddleware(
			MultiReplaceAttribute(FormatError, NewLevels)),
	}
}
