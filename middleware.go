package zlog

import (
	"context"
	"log/slog"
)

// Middleware is the unified pipeline type that replaces both modifiers and replaceAttr functions.
type Middleware func(ctx context.Context, r slog.Record) (context.Context, slog.Record)

func MultiReplaceAttribute(fns ...func(slog.Attr) slog.Attr) func(slog.Attr) slog.Attr {
	return func(a slog.Attr) slog.Attr {
		for _, fn := range fns {
			a = fn(a)
		}
		return a
	}
}

// ReplaceAttributeMiddleware adapts a per-attribute transform function into a Middleware.
func ReplaceAttributeMiddleware(fn func(slog.Attr) slog.Attr) Middleware {
	return func(ctx context.Context, r slog.Record) (context.Context, slog.Record) {
		newR := slog.NewRecord(r.Time, r.Level, r.Message, r.PC)
		r.Attrs(func(a slog.Attr) bool {
			newR.AddAttrs(fn(a))
			return true
		})
		return ctx, newR
	}
}

type slogctxattrs struct{}

// ContextMiddleware pulls slog attributes stored in the context and injects them into the record.
// Use AddToContext to attach attributes to a context.
func ContextMiddleware(ctx context.Context, r slog.Record) (context.Context, slog.Record) {
	if val := ctx.Value(slogctxattrs{}); val != nil {
		if attrs, ok := val.([]slog.Attr); ok {
			r.AddAttrs(attrs...)
		}
	}
	return ctx, r
}

// AddToContext stores slog attributes in the context for injection by ContextMiddleware.
func AddToContext(ctx context.Context, attrs ...slog.Attr) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	var ctxattrs []slog.Attr
	if val := ctx.Value(slogctxattrs{}); val != nil {
		if old, ok := val.([]slog.Attr); ok {
			ctxattrs = old
		}
	}

	return context.WithValue(ctx, slogctxattrs{}, append(ctxattrs, attrs...))
}
