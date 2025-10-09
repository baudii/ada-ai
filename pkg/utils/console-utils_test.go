package utils

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type errorReader struct {
	r   *strings.Reader
	cnt *int
}

func (e errorReader) Read(p []byte) (int, error) {
	if *e.cnt == 0 {
		*e.cnt = 1
		return 0, fmt.Errorf("simulated read error")
	}

	return e.r.Read(p)
}

func TestReadInputInternal(t *testing.T) {
	origrd := reader
	origwr := writer
	origlg := logger

	t.Cleanup(func() {
		reader = origrd
		writer = origwr
		logger = origlg
	})

	var cnt = 0
	tests := []struct {
		r         io.Reader
		w         io.Writer
		logwriter *bytes.Buffer
		msg       string
		expRead   string
		expWr     string
		logsErr   bool
		panics    bool
	}{
		{strings.NewReader("a"), &bytes.Buffer{}, &bytes.Buffer{}, "i", "a", "i > ", false, false},
		{strings.NewReader("a\na"), &bytes.Buffer{}, &bytes.Buffer{}, "", "a", " > ", false, false},
		{strings.NewReader("abc\nabc\nabc"), &bytes.Buffer{}, &bytes.Buffer{}, "abc", "abc", "abc > ", false, false},
		{errorReader{r: strings.NewReader("f1"), cnt: &cnt}, &bytes.Buffer{}, &bytes.Buffer{}, "---", "f1", "--- > --- > ", true, false},
		{errorReader{r: strings.NewReader("f1"), cnt: &cnt}, &bytes.Buffer{}, &bytes.Buffer{}, "---", "f1", "--- > --- > ", true, true},
	}

	for i, v := range tests {
		reader = v.r
		writer = v.w
		logger = slog.New(slog.NewTextHandler(v.logwriter, nil))
		if v.panics {
			writer = &mockWriter{func() (int, error) { return 0, fmt.Errorf("failed") }}
			assert.Panics(t, func() { _ = ReadInput(v.msg) }, "test: %v", i)
			continue
		}
		res := ReadInput(v.msg)

		if !v.logsErr {
			assert.True(t, len(v.logwriter.String()) == 0, "test: %v", i)
		} else {
			assert.True(t, len(v.logwriter.String()) > 0, "test: %v", i)
		}

		b, ok := v.w.(*bytes.Buffer)
		if assert.True(t, ok, "test: %v", i) {
			assert.Equal(t, v.expWr, b.String(), "test: %v", i)
		}
		assert.Equal(t, v.expRead, res, "test: %v", i)
	}
}
