package dilog

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	getDate func(*time.Location) string = func(loc *time.Location) string {
		return time.Now().In(loc).Format("2006-01-02")
	}
	getFileName func(*dailyWriter, string) string = func(dw *dailyWriter, date string) string {
		return filepath.Join(dw.path, fmt.Sprintf("%s_%s.log", dw.prefix, date))
	}
)

type dailyWriter struct {
	path     string
	prefix   string
	timezone *time.Location

	mu      sync.Mutex
	curDate string
	file    *os.File
}

// Stores global instance of *dailyWriter. Helpful when you need
// to write something directly into the log file. Otherwise, it is
// preffered to use slog.Logger methods.
var Dw *dailyWriter

// Provides parallel-safe writing functionality to the *os.File stored
// in dw.file. Performs a rotation by calling rotateIfNeeded() method
// which rotates the target file if new date is different from dw.curDate.
// Basically, this means that file is rotated every day according to dw.timezone.
func (dw *dailyWriter) Write(p []byte) (n int, err error) {
	dw.mu.Lock()
	defer dw.mu.Unlock()
	if err := dw.rotateIfNeeded(); err != nil {
		return 0, err
	}
	return dw.file.Write(p)
}

func (dw *dailyWriter) rotateIfNeeded() error {
	date := getDate(dw.timezone)
	if dw.file != nil && date == dw.curDate {
		return nil
	}

	if dw.file != nil {
		_ = dw.file.Close()
		dw.file = nil
	}

	if err := os.MkdirAll(dw.path, 0o755); err != nil {
		return fmt.Errorf("rotate: %w", err)
	}

	f, err := os.OpenFile(getFileName(dw, date), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o666)
	if err != nil {
		return fmt.Errorf("rotate: %w", err)
	}

	dw.file = f
	dw.curDate = date
	return nil
}
