package log

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/rs/zerolog"
)

type Config struct {
	Writers []io.Writer
	Level   zerolog.Level
	WithF   func(zerolog.Context) zerolog.Logger
	CallerF func(pc uintptr, file string, line int) string

	once sync.Once
	lg   zerolog.Logger
}

func (c *Config) applyDefaults() {
	if c.Writers == nil {
		c.Writers = defaultConfig.Writers
	}
	if c.Level == 0 {
		c.Level = defaultConfig.Level
	}
	if c.WithF == nil {
		c.WithF = defaultConfig.WithF
	}
	if c.CallerF == nil {
		c.CallerF = defaultConfig.CallerF
	}
}

func (c *Config) logger() zerolog.Logger {
	c.once.Do(func() {
		c.applyDefaults()
		zerolog.CallerMarshalFunc = c.CallerF
		c.lg = c.WithF(zerolog.New(zerolog.MultiLevelWriter(c.Writers...)).Level(c.Level).With())
	})

	return c.lg
}

var (
	defaultConfig = Config{
		Writers: []io.Writer{zerolog.ConsoleWriter{Out: os.Stdout}},
		Level:   zerolog.DebugLevel,
		WithF:   defaultWithF,
		CallerF: defaultCallerF,
	}
	currentConfig = struct {
		mu     sync.RWMutex
		config *Config
	}{config: &defaultConfig}
)

func SetConfig(config *Config) {
	currentConfig.mu.Lock()
	defer currentConfig.mu.Unlock()
	currentConfig.config = config
}

func defaultWithF(zc zerolog.Context) zerolog.Logger {
	return zc.Timestamp().
		Caller().
		Logger()
}

func defaultCallerF(pc uintptr, file string, line int) string {
	parts := strings.Split(runtime.FuncForPC(pc).Name(), "/")
	pkgparts := strings.Split(parts[len(parts)-1], ".")
	if len(parts) == 1 {
		pkg := pkgparts[0]
		return fmt.Sprintf("%s:%sL%d", pkg, filepath.Base(file), line)
	}

	parent, pkg := parts[len(parts)-2], pkgparts[0]

	return fmt.Sprintf("%s/%s:%sL%d", parent, pkg, filepath.Base(file), line)
}

func Logger() *zerolog.Logger {
	currentConfig.mu.RLock()
	defer currentConfig.mu.RUnlock()

	return new(currentConfig.config.logger())
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
