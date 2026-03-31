package log

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/rs/zerolog"
)

var (
	logger     zerolog.Logger
	loggerOnce = sync.Once{}
)

func Logger() *zerolog.Logger {
	loggerOnce.Do(func() {
		zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
			parts := strings.Split(runtime.FuncForPC(pc).Name(), "/")
			pkgparts := strings.Split(parts[len(parts)-1], ".")
			if len(parts) == 1 {
				pkg := pkgparts[0]
				return fmt.Sprintf("%s:%sL%d", pkg, filepath.Base(file), line)
			}

			parent, pkg := parts[len(parts)-2], pkgparts[0]

			return fmt.Sprintf("%s/%s:%sL%d", parent, pkg, filepath.Base(file), line)
		}

		logger = zerolog.New(zerolog.MultiLevelWriter(
			zerolog.ConsoleWriter{Out: os.Stdout},
			// os.Stderr,
		)).
			Level(zerolog.TraceLevel).
			With().
			Timestamp().
			Caller().
			Logger()
	})

	return &logger
}

func Ctx(ctx context.Context) *zerolog.Logger {
	l := zerolog.Ctx(ctx)
	if l.GetLevel() == zerolog.Disabled {
		return Logger()
	}
	return l
}

func Trace(ctx context.Context) *zerolog.Event {
	return Ctx(ctx).Trace()
}

func Debug(ctx context.Context) *zerolog.Event {
	return Ctx(ctx).Debug()
}

func Info(ctx context.Context) *zerolog.Event {
	return Ctx(ctx).Info()
}

func Warn(ctx context.Context) *zerolog.Event {
	return Ctx(ctx).Warn()
}

func Error(ctx context.Context) *zerolog.Event {
	return Ctx(ctx).Error()
}

func Panic(ctx context.Context) *zerolog.Event {
	return Ctx(ctx).Panic()
}

func Fatal(ctx context.Context) *zerolog.Event {
	return Ctx(ctx).Fatal()
}

func WithCtx(ctx context.Context, l zerolog.Logger) context.Context {
	return l.WithContext(ctx)
}

func Span(ctx context.Context, span string) context.Context {
	return Ctx(ctx).With().Str("span", span).Logger().WithContext(ctx)
}
