package zlog

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/fatih/color"
)

var (
	//
	colorBody    = color.New(color.FgHiWhite, color.Italic)
	colorMessage = color.RGB(170, 170, 170)
	colorTimestr = color.New(color.FgCyan, color.Italic)
)

func MultiReplaceAttr(fns ...func([]string, slog.Attr) slog.Attr) func([]string, slog.Attr) slog.Attr {
	return func(groups []string, a slog.Attr) slog.Attr {
		for _, fn := range fns {
			a = fn(groups, a)
		}

		return a
	}
}

func FormatError(_ []string, a slog.Attr) slog.Attr {
	switch x := a.Value.Any().(type) {
	case error:
		a = slog.GroupAttrs(a.Key, formatError(x)...)
	}

	return a
}

type formattedHandler struct {
	out         io.Writer
	lowest      slog.Level
	attrs       []slog.Attr
	groups      []string
	modifiers   []modifier
	replaceAttr func([]string, slog.Attr) slog.Attr
}

func (h formattedHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= getLevel(ctx, h.lowest)
}

func (h formattedHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &formattedHandler{
		out:       h.out,
		lowest:    h.lowest,
		attrs:     append(h.attrs, attrs...),
		groups:    h.groups,
		modifiers: h.modifiers,
	}
}

func (h formattedHandler) WithGroup(name string) slog.Handler {
	return &formattedHandler{
		out:       h.out,
		lowest:    h.lowest,
		attrs:     h.attrs,
		groups:    append(h.groups, name),
		modifiers: h.modifiers,
	}
}

func (h formattedHandler) applyModifiers(ctx context.Context, r slog.Record) (context.Context, slog.Record) {
	for _, m := range h.modifiers {
		ctx, r = m(ctx, r)
	}

	return ctx, r
}

func (h formattedHandler) attrsToMap(r slog.Record) map[string]any {
	fields := make(map[string]any, r.NumAttrs())
	r.Attrs(func(a slog.Attr) bool {
		fields[a.Key] = a.Value.Any()

		return true
	})

	return fields
}

func (h formattedHandler) Handle(ctx context.Context, r slog.Record) error {
	defer executeLevelSpecificActions(r)

	r.AddAttrs(h.attrs...)
	ctx, r = h.applyModifiers(ctx, r)
	level := formatLevel(r)
	fields := h.attrsToMap(r)

	bodyBytes, err := json.MarshalIndent(fields, "", "  ")
	if err != nil {
		return err
	}

	body := colorBody.Sprint(string(bodyBytes))
	timeStr := colorTimestr.Sprint(r.Time.Format("[_2/01 15:04:05Z07]"))
	msg := colorMessage.Sprint(r.Message)

	fmt.Fprintln(h.out, timeStr, level, msg, body)

	return nil
}

func New(minLvl ...slog.Level) *slog.Logger {
	lvl := LevelDebug
	if len(minLvl) != 0 {
		lvl = minLvl[0]
	}

	return slog.New(&formattedHandler{
		out:         os.Stdout,
		lowest:      lvl,
		modifiers:   []modifier{ContextModifier},
		replaceAttr: MultiReplaceAttr(NewLevels, FormatError),
	})

}
