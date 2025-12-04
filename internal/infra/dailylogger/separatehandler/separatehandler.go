package separatehandler

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"
)

// LevelWriter associates an io.Writer with a specific log level.
type LevelWriter struct {
	W     io.Writer
	Level slog.Level
}

// SeparateHandler is a basic implementation of slog.Handler that writes log records to an io.Writer.
type SeparateHandler struct {
	levelWriters []LevelWriter
	attrs        []slog.Attr
	loc          *time.Location
}

// Option configures a SimpleHandler instance.
type Option func(*SeparateHandler)

// WithLevelWriters sets the level-specific writers for the SeparateHandler. It expects pairs of slog.Level and io.Writer.
// For example: WithLevelWriters(slog.LevelInfo, infoWriter, slog.LevelError, errorWriter).
func WithLevelWriters(writers ...any) Option {
	return func(h *SeparateHandler) {
		h.levelWriters = []LevelWriter{}
		for i := 0; i < len(writers)-1; i += 2 {
			level, ok := writers[i].(slog.Level)
			if !ok {
				continue
			}
			writer, ok := writers[i+1].(io.Writer)
			if !ok {
				continue
			}
			h.levelWriters = append(h.levelWriters, LevelWriter{W: writer, Level: level})
		}
	}
}

// WithLocation sets the time location for the SimpleHandler.
func WithLocation(loc *time.Location) Option {
	return func(h *SeparateHandler) {
		h.loc = loc
	}
}

// WithAttrs adds attributes to the handler.
func (h *SeparateHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	h.attrs = append(h.attrs, attrs...)
	return h
}

// WithGroup is a no-op for SimpleHandler.
func (h *SeparateHandler) WithGroup(name string) slog.Handler {
	return h
}

// NewSimpleHandler creates a new SimpleHandler that writes to the provided io.Writer.
func New(opts ...Option) *SeparateHandler {
	h := &SeparateHandler{loc: time.UTC}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

// Enabled checks if the given log level is enabled.
func (h *SeparateHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, aw := range h.levelWriters {
		if level >= aw.Level {
			return true
		}
	}
	return false
}

// Handle formats and writes the log record to the io.Writer. It includes the time, level, message, and any attributes.
// The time is formatted according to the handler's location.
func (h *SeparateHandler) Handle(ctx context.Context, r slog.Record) error {
	time := r.Time.In(h.loc).Format("2006-01-02 15:04:05")
	level := r.Level.String()
	msg := r.Message

	var attrs strings.Builder
	for _, a := range h.attrs {
		attrs.WriteString(fmt.Sprintf(" %s=%v", a.Key, a.Value))
	}
	r.Attrs(func(a slog.Attr) bool {
		attrs.WriteString(fmt.Sprintf(" %s=%v", a.Key, a.Value))
		return true
	})

	line := fmt.Sprintf("%s %-5s %s%s\n", time, level, msg, attrs.String())
	for _, levelWriter := range h.levelWriters {
		if r.Level >= levelWriter.Level {
			_, err := levelWriter.W.Write([]byte(line))
			if err != nil {
				return err
			}
		}
	}

	return nil
}
