package utils

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
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

	var cnt int = 0
	tests := []struct {
		r         io.Reader
		w         *bytes.Buffer
		logwriter *bytes.Buffer
		msg       string
		expRead   string
		expWr     string
		logsErr   bool
	}{
		{strings.NewReader("a"), &bytes.Buffer{}, &bytes.Buffer{}, "i", "a", "i > ", false},
		{strings.NewReader("a\na"), &bytes.Buffer{}, &bytes.Buffer{}, "", "a", " > ", false},
		{strings.NewReader("abc\nabc\nabc"), &bytes.Buffer{}, &bytes.Buffer{}, "abc", "abc", "abc > ", false},
		{errorReader{r: strings.NewReader("f1"), cnt: &cnt}, &bytes.Buffer{}, &bytes.Buffer{}, "---", "f1", "--- > --- > ", true},
	}

	for _, v := range tests {
		reader = v.r
		writer = v.w
		logger = slog.New(slog.NewTextHandler(v.logwriter, nil))
		res := ReadInput(v.msg)

		if !v.logsErr {
			require.True(t, len(v.logwriter.String()) == 0)
		} else {
			require.True(t, len(v.logwriter.String()) > 0)
		}

		require.Equal(t, v.expWr, v.w.String())
		require.Equal(t, v.expRead, res)
	}
}
