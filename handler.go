package zlog

import (
	"context"
	"fmt"
	"io"
	"log/slog"
)

type handler struct {
	out        io.Writer
	lowest     slog.Level
	attrs      []slog.Attr
	groups     []string
	middleware []Middleware
	style      Style
}

func newHandler(out io.Writer, lowest slog.Level, style Style, mw []Middleware) *handler {
	mwCopy := make([]Middleware, len(mw))
	copy(mwCopy, mw)
	return &handler{
		out:        out,
		lowest:     lowest,
		middleware: mwCopy,
		style:      style,
	}
}

func (h *handler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= getLevel(ctx, h.lowest)
}

func (h *handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	cp := *h
	cp.attrs = append(append([]slog.Attr(nil), h.attrs...), attrs...)
	return &cp
}

func (h *handler) WithGroup(name string) slog.Handler {
	cp := *h
	cp.groups = append(append([]string(nil), h.groups...), name)
	return &cp
}

func (h *handler) Handle(ctx context.Context, r slog.Record) error {
	r.AddAttrs(h.attrs...)
	ctx, r = applyMiddleware(ctx, r, h.middleware)
	fields := attrsToMap(r)

	line, err := h.style.render(r, fields)
	if err != nil {
		return err
	}

	fmt.Fprintln(h.out, line)
	executeLevelSpecificActions(r)
	return nil
}

func applyMiddleware(ctx context.Context, r slog.Record, mw []Middleware) (context.Context, slog.Record) {
	for _, m := range mw {
		ctx, r = m(ctx, r)
	}
	return ctx, r
}

func attrsToMap(r slog.Record) map[string]any {
	fields := make(map[string]any, r.NumAttrs())
	r.Attrs(func(a slog.Attr) bool {
		fields[a.Key] = attrToValue(a.Value)
		return true
	})
	return fields
}

func attrToValue(v slog.Value) any {
	v = v.Resolve()
	switch v.Kind() {
	case slog.KindGroup:
		m := make(map[string]any, len(v.Group()))
		for _, a := range v.Group() {
			m[a.Key] = attrToValue(a.Value)
		}
		return m
	case slog.KindString:
		return v.String()
	case slog.KindInt64:
		return v.Int64()
	case slog.KindUint64:
		return v.Uint64()
	case slog.KindFloat64:
		return v.Float64()
	case slog.KindBool:
		return v.Bool()
	case slog.KindTime:
		return v.Time()
	case slog.KindDuration:
		return v.Duration()
	default:
		return v.Any()
	}
}

// multiHandler fans log records out to multiple slog.Handler instances.
type multiHandler struct {
	handlers []slog.Handler
}

func (m *multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range m.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (m *multiHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, h := range m.handlers {
		if h.Enabled(ctx, r.Level) {
			if err := h.Handle(ctx, r.Clone()); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		handlers[i] = h.WithAttrs(attrs)
	}
	return &multiHandler{handlers: handlers}
}

func (m *multiHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		handlers[i] = h.WithGroup(name)
	}
	return &multiHandler{handlers: handlers}
}
