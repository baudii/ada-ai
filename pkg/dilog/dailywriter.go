package dilog

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// LogConfig holds configuration settings for logging.
type LogConfig struct {
	Timezone string `json:"timezone"`
	Path     string `json:"path"`
	Prefix   string `json:"prefix"`
	Level    string `json:"level"`
}

type dailyWriter struct {
	path   string
	prefix string
	loc    *time.Location

	date    dater
	prepare preparer
	open    opener

	mu      sync.Mutex
	curDate string
	wc      io.WriteCloser
}

type Option func(*dailyWriter)

type (
	dater    func(loc *time.Location) string
	opener   func(path, prefix, date string) (io.WriteCloser, error)
	preparer func(path string) error
)

var (
	getDate = func(loc *time.Location) string {
		return time.Now().In(loc).Format("2006-01-02")
	}
	open = func(path, prefix, date string) (io.WriteCloser, error) {
		f := filepath.Join(path, fmt.Sprintf("%s_%s.log", prefix, date))
		return os.OpenFile(f, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o666)
	}
	prepare = func(path string) error {
		return os.MkdirAll(path, 0o755)
	}
)

// WithPrefix sets the prefix for the log files created by the dailyWriter.
func WithPrefix(prefix string) Option {
	return func(dw *dailyWriter) {
		dw.prefix = prefix
	}
}

// WithLocation sets the time location for the dailyWriter.
func WithLocation(loc *time.Location) Option {
	return func(dw *dailyWriter) {
		dw.loc = loc
	}
}

// WithDater sets the date format function for the dailyWriter.
func WithDater(dg dater) Option {
	return func(dw *dailyWriter) {
		dw.date = dg
	}
}

// WithOpener sets the file opener function for the dailyWriter.
func WithOpener(o opener) Option {
	return func(dw *dailyWriter) {
		dw.open = o
	}
}

// WithPreparer sets the preparation function for the dailyWriter.
func WithPreparer(m preparer) Option {
	return func(dw *dailyWriter) {
		dw.prepare = m
	}
}

// NewDailyWriter creates and returns a new dailyWriter instance based on the provided Config.
// The dailyWriter manages log file rotation based on the date and writes log entries to the appropriate file.
func NewDailyWriter(path string, opts ...Option) (*dailyWriter, error) {
	dw := &dailyWriter{
		path:    path,
		prefix:  "default",
		loc:     time.UTC,
		date:    getDate,
		open:    open,
		prepare: prepare,
	}

	for _, opt := range opts {
		opt(dw)
	}

	if err := dw.tryRotate(); err != nil {
		return nil, fmt.Errorf("new dailywriter: %w", err)
	}

	return dw, nil
}

// Write provides parallel-safe writing functionality to the *os.File stored
// in dw.file. Performs a rotation by calling rotateIfNeeded() method
// which rotates the target file if new date is different from dw.curDate.
// Basically, this means that file is rotated every day according to dw.timezone.
func (dw *dailyWriter) Write(p []byte) (n int, err error) {
	dw.mu.Lock()
	defer dw.mu.Unlock()
	if err := dw.tryRotate(); err != nil {
		return 0, err
	}
	return dw.wc.Write(p)
}

func (dw *dailyWriter) tryRotate() error {
	date := dw.date(dw.loc)
	if dw.wc != nil && dw.curDate == date {
		return nil
	}

	if dw.wc != nil {
		_ = dw.wc.Close()
		dw.wc = nil
	}

	if err := dw.prepare(dw.path); err != nil {
		return fmt.Errorf("rotate prepare: %w", err)
	}

	wc, err := dw.open(dw.path, dw.prefix, date)
	if err != nil {
		return fmt.Errorf("rotate open: %w", err)
	}

	dw.wc = wc
	dw.curDate = date
	return nil
}
