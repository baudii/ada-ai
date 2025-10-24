package simplehandler

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"
)

// SimpleHandler is a basic implementation of slog.Handler that writes log records to an io.Writer.
type SimpleHandler struct {
    w     io.Writer
    attrs []slog.Attr
    level slog.Level
    loc   *time.Location
}

// Option configures a SimpleHandler instance.
type Option func(*SimpleHandler)

// WithLevel sets the log level for the SimpleHandler.
func WithLevel(level slog.Level) Option {
	return func(h *SimpleHandler) {
		h.level = level
	}
}

// WithLocation sets the time location for the SimpleHandler.
func WithLocation(loc *time.Location) Option {
	return func(h *SimpleHandler) {
		h.loc = loc
	}
}

// WithAttrs adds attributes to the handler.
func (h *SimpleHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	h.attrs = append(h.attrs, attrs...)
	return h
}

// WithGroup is a no-op for SimpleHandler.
func (h *SimpleHandler) WithGroup(name string) slog.Handler {
	return h
}

// NewSimpleHandler creates a new SimpleHandler that writes to the provided io.Writer.
func NewSimpleHandler(w io.Writer, opts ...Option) *SimpleHandler {
	h := &SimpleHandler{w: w, level: slog.LevelInfo, loc: time.UTC}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

// Enabled checks if the given log level is enabled.
func (h *SimpleHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.level
}

// Handle formats and writes the log record to the io.Writer. It includes the time, level, message, and any attributes.
// The time is formatted according to the handler's location.
func (h *SimpleHandler) Handle(ctx context.Context, r slog.Record) error {
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
	_, err := h.w.Write([]byte(line))
	return err
}
