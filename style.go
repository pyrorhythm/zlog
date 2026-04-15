package zlog

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"maps"
	"time"

	"github.com/fatih/color"
)

// Format controls how log lines are rendered.
type Format int

const (
	FormatPretty Format = iota
	FormatJSON
)

// ColorScheme defines the color palette.
type ColorScheme struct {
	Levels  map[slog.Level]*color.Color
	Time    *color.Color
	Message *color.Color
	Body    *color.Color
}

// Style controls the visual presentation of a log sink.
type Style struct {
	Format     Format
	TimeFormat string
	Colors     ColorScheme
}

var DefaultStyle = Style{
	Format:     FormatPretty,
	TimeFormat: "[_2/01 15:04:05Z07]",
	Colors: ColorScheme{
		Levels: map[slog.Level]*color.Color{
			LevelTrace: color.RGB(67, 120, 198),
			LevelDebug: color.RGB(65, 185, 198),
			LevelInfo:  color.RGB(134, 180, 249),
			LevelWarn:  color.RGB(252, 172, 75),
			LevelError: color.RGB(252, 99, 75),
			LevelPanic: color.RGB(216, 52, 0),
			LevelFatal: color.RGB(73, 0, 0),
		},
		Time:    color.New(color.FgCyan, color.Italic),
		Message: color.RGB(170, 170, 170),
		Body:    color.New(color.FgHiWhite, color.Italic),
	},
}

var JSONStyle = Style{
	Format:     FormatJSON,
	TimeFormat: time.RFC3339,
}

func (s Style) render(r slog.Record, fields map[string]any) (string, error) {
	switch s.Format {
	case FormatJSON:
		return s.renderJSON(r, fields)
	default:
		return s.renderPretty(r, fields)
	}
}

func (s Style) renderPretty(r slog.Record, fields map[string]any) (string, error) {
	cs := s.Colors

	timeStr := r.Time.Format(s.TimeFormat)
	if cs.Time != nil {
		timeStr = cs.Time.Sprint(timeStr)
	}

	level := s.levelString(r.Level)

	msg := r.Message
	if cs.Message != nil {
		msg = cs.Message.Sprint(msg)
	}

	if len(fields) == 0 {
		return fmt.Sprintf("%s %s %s", timeStr, level, msg), nil
	}

	bodyBytes, err := json.MarshalIndent(fields, "", "  ")
	if err != nil {
		return "", err
	}

	body := string(bodyBytes)
	if cs.Body != nil {
		body = cs.Body.Sprint(body)
	}

	return fmt.Sprintf("%s %s %s %s", timeStr, level, msg, body), nil
}

func (s Style) renderJSON(r slog.Record, fields map[string]any) (string, error) {
	entry := make(map[string]any, len(fields)+3)
	maps.Copy(entry, fields)
	entry["time"] = r.Time.Format(s.TimeFormat)
	entry["level"] = levelName(r.Level)
	entry["msg"] = r.Message

	b, err := json.Marshal(entry)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (s Style) levelString(level slog.Level) string {
	str := levelName(level)
	if s.Format == FormatPretty {
		if col, ok := s.Colors.Levels[level]; ok {
			return col.Sprint(str)
		}
	}
	return str
}
