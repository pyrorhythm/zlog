package zlog

import (
	"context"
	"log/slog"
)

type ctxLevelKey struct{}

func SetLevel(ctx context.Context, level slog.Level) context.Context {
	return context.WithValue(ctx, ctxLevelKey{}, level)
}

func getLevel(ctx context.Context, fallback slog.Level) slog.Level {
	lvl := ctx.Value(ctxLevelKey{})
	if lvl == nil {
		return fallback
	}
	slvl, ok := lvl.(slog.Level)
	if !ok {
		return fallback
	}
	return slvl
}
