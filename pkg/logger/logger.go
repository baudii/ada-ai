package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/baudii/ada-ai/pkg/utils"
)

type DailyWriter struct {
	mu      sync.Mutex
	dir     string
	prefix  string
	loc     *time.Location
	curDate string
	file    *os.File
}

func (dw *DailyWriter) Write(p []byte) (n int, err error) {
	dw.mu.Lock()
	defer dw.mu.Unlock()
	if err := dw.rotateIfNeeded(); err != nil {
		return 0, err
	}
	return dw.file.Write(p)
}

func (dw *DailyWriter) filename(t time.Time) string {
	return filepath.Join(dw.dir, fmt.Sprintf("%s_%s.log", dw.prefix, t.Format("2006-01-02")))
}

func (dw *DailyWriter) rotateIfNeeded() error {
	now := time.Now().In(dw.loc)
	date := now.Format("2006-01-02")
	if dw.file != nil && date == dw.curDate {
		return nil
	}

	if dw.file != nil {
		_ = dw.file.Close()
		dw.file = nil
	}

	if err := os.MkdirAll(dw.dir, 0o755); err != nil {
		return err
	}

	f, err := os.OpenFile(dw.filename(now), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}

	dw.file = f
	dw.curDate = date
	return nil
}

func New() error {
	dwWriter := &DailyWriter{
		loc:    time.Local,
		dir:    utils.GetAbsolutePath("logs"),
		prefix: "ada",
	}

	if err := dwWriter.rotateIfNeeded(); err != nil {
		return err
	}

	writer := io.MultiWriter(os.Stdout, dwWriter)
	lgr := slog.New(slog.NewTextHandler(writer, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(lgr)
	return nil
}
