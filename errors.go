package zlog

import (
	"log/slog"
	"runtime"

	"github.com/pkg/errors"
)

type frame struct {
	Function string `json:"function"`
	Source   string `json:"source"`
	Line     int    `json:"line"`
}

func (f *frame) convert(p errors.Frame) *frame {
	pc := uintptr(p) - 1
	function := runtime.FuncForPC(pc)
	functionName := "<unknown>"
	file, line := "<unknown>", -1
	if function != nil {
		functionName = function.Name()
		file, line = function.FileLine(pc)
	}

	*f = frame{
		Function: functionName,
		Line:     line,
		Source:   file,
	}

	return f
}

// FormatError is a replaceAttr-compatible function that expands error attributes
// into a group with msg, cause, and stack frames.
func FormatError(a slog.Attr) slog.Attr {
	if x, ok := a.Value.Any().(error); ok {
		a = slog.GroupAttrs(a.Key, formatError(x)...)
	}
	return a
}

func formatError(err error) []slog.Attr {
	var errGroup []slog.Attr

	if err == nil {
		return errGroup
	}

	errGroup = append(errGroup,
		slog.String("msg", err.Error()))

	if cerr, ok := err.(interface {
		Cause() error
	}); ok {
		errGroup = append(errGroup,
			slog.Any("cause", cerr.Cause()))
	}

	if sterr, ok := err.(interface {
		StackTrace() errors.StackTrace
	}); ok {
		stackTrace := sterr.StackTrace()
		frames := make([]*frame, len(stackTrace))

		for i, fr := range stackTrace {
			frames[i] = new(frame).convert(fr)
		}

		errGroup = append(errGroup,
			slog.Any("frames", frames))
	}

	return errGroup
}
