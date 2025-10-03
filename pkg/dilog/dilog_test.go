package dilog

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	today := "2025-10-05"
	origW, origF, origGf := w, getDate, getFileName
	getDate = func(*time.Location) string { return today }
	t.Cleanup(func() {
		w = origW
		getDate = origF
		getFileName = origGf
		Dw.file.Close()
		Dw = nil
	})

	tests := []struct {
		cfg     *Config
		level   slog.Level
		curDate string
		logFunc func(string, ...any)
		w       io.Writer
		gf      func(*dailyWriter, string) string
	}{
		{&Config{"Local", t.TempDir(), "test_log", "debug"}, slog.LevelDebug, today, slog.Debug, &bytes.Buffer{}, nil},
		{&Config{"UTC", t.TempDir(), "test_log", "info"}, slog.LevelInfo, "2024-10-04", slog.Info, &bytes.Buffer{}, nil},
		{&Config{"America/New_York", t.TempDir(), "test_log", "warn"}, slog.LevelWarn, "", slog.Warn, &bytes.Buffer{}, func(dw *dailyWriter, date string) string { return "" }},
		{&Config{"Invalid/Timezone", filepath.Join(t.TempDir(), "t2.txt"), "test_log", "error"}, slog.LevelError, "2024-10-03", slog.Error, &bytes.Buffer{}, nil},
	}

	for _, v := range tests {
		w = v.w
		if v.curDate != "" {
			path := filepath.Join(v.cfg.Path, "random.log")
			if strings.HasSuffix(v.cfg.Path, "t1.txt") {
				path = v.cfg.Path
				v.cfg.Path = filepath.Dir(v.cfg.Path)
			} else if strings.HasSuffix(v.cfg.Path, "t2.txt") {
				path = v.cfg.Path
			}
			f, err := os.Create(path)
			require.NoError(t, err)
			getDw = func(c *Config, l *time.Location) *dailyWriter {
				return &dailyWriter{
					path:     c.Path,
					prefix:   c.Prefix,
					timezone: l,
					file:     f,
					curDate:  v.curDate,
				}
			}
		}

		if v.gf != nil {
			getFileName = v.gf
		}
		Init(v.cfg)
		if v.curDate == today {
			Dw.curDate = "2025-10-25"
			Dw.path = ""
		}

		func(l slog.Level, log func(string, ...any)) {
			if slog.Default().Enabled(context.Background(), l) {
				log("Message")
				assert.Contains(t, w.(*bytes.Buffer).String(), "Message")
			}
		}(v.level, v.logFunc)
		Dw.file.Close()
	}
}

func TestWritelnToDw(t *testing.T) {
	origDw := Dw
	origGt := getDate
	getDate = func(*time.Location) string { return "2025-10-05" }
	t.Cleanup(func() {
		Dw = origDw
		getDate = origGt
	})
	path := filepath.Join(t.TempDir(), "test.log")
	f, err := os.Create(path)
	require.NoError(t, err)
	Dw = &dailyWriter{
		file:     f,
		path:     filepath.Dir(path),
		timezone: time.UTC,
		prefix:   "test",
		curDate:  "2025-10-05",
	}
	WritelnToDw("Test message")
	require.FileExists(t, path)
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "Test message\n", string(content))
	Dw.file.Close()
}
