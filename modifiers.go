package zlog

import (
	"context"
	"log/slog"
)

type modifier func(ctx context.Context, r slog.Record) (context.Context, slog.Record)
type slogctxattrs struct{}

func ContextModifier(ctx context.Context, r slog.Record) (context.Context, slog.Record) {
	if val := ctx.Value(slogctxattrs{}); val != nil {
		attrs, ok := val.([]slog.Attr)
		if ok {
			r.AddAttrs(attrs...)
		}
	}

	return ctx, r
}

func AddToContext(ctx context.Context, attrs ...slog.Attr) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	var ctxattrs []slog.Attr

	if val := ctx.Value(slogctxattrs{}); val != nil {
		if oldattrs, ok := val.([]slog.Attr); ok {
			ctxattrs = oldattrs
		}
	}

	ctxattrs = append(ctxattrs, attrs...)

	return context.WithValue(ctx, slogctxattrs{}, ctxattrs)
}
