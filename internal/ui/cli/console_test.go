package cli

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockWriter struct {
	f func() (int, error)
}

func (m *mockWriter) Write(p []byte) (n int, err error) {
	return m.f()
}

type mockReader struct {
	r   *strings.Reader
	cnt *int
}

func (e mockReader) Read(p []byte) (int, error) {
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
		{
			strings.NewReader("a"),
			&bytes.Buffer{},
			&bytes.Buffer{},
			"i",
			"a",
			"i > ",
			false,
			false,
		},
		{
			strings.NewReader("a\na"),
			&bytes.Buffer{},
			&bytes.Buffer{},
			"",
			"a",
			" > ",
			false,
			false,
		},
		{
			strings.NewReader("abc\nabc\nabc"),
			&bytes.Buffer{},
			&bytes.Buffer{},
			"abc",
			"abc",
			"abc > ",
			false,
			false,
		},
		{
			mockReader{r: strings.NewReader("f1"), cnt: &cnt},
			&bytes.Buffer{},
			&bytes.Buffer{},
			"---",
			"f1",
			"--- > --- > ",
			true,
			false,
		},
		{
			mockReader{r: strings.NewReader("f1"), cnt: &cnt},
			&bytes.Buffer{},
			&bytes.Buffer{},
			"---",
			"f1",
			"--- > --- > ",
			true,
			true,
		},
	}

	for i, v := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			reader = v.r
			writer = v.w
			logger = slog.New(slog.NewTextHandler(v.logwriter, nil))
			if v.panics {
				writer = &mockWriter{func() (int, error) { return 0, fmt.Errorf("failed") }}
				assert.Panics(t, func() { _ = ReadInput(v.msg) })
				return
			}
			res := ReadInput(v.msg)

			if !v.logsErr {
				assert.True(t, len(v.logwriter.String()) == 0)
			} else {
				assert.True(t, len(v.logwriter.String()) > 0)
			}

			b, ok := v.w.(*bytes.Buffer)
			if assert.True(t, ok) {
				assert.Equal(t, v.expWr, b.String())
			}
			assert.Equal(t, v.expRead, res)
		})
	}
}

func TestPrintTree(t *testing.T) {
	t.Parallel()
	tests := []struct {
		m        map[string]any
		expected []string
		hasErr   bool
	}{
		{map[string]any{"a": map[string]any{"b": 0}}, []string{"└── a\n    └── b\n"}, false},
		{map[string]any{"a": map[string]any{"b": 0, "c": 0}, "b": map[string]any{"c": 0}}, []string{
			"├── a\n│   ├── c\n│   └── b\n└── b\n    └── c\n",
			"├── a\n│   ├── b\n│   └── c\n└── b\n    └── c\n",
			"├── b\n│   └── c\n└── a\n    ├── b\n    └── c\n",
			"├── b\n│   └── c\n└── a\n    ├── c\n    └── b\n",
		}, false},
		{map[string]any{"a": map[string]any{"b": 0, "c": 0}, "b": map[string]any{"c": 0}}, nil, true},
	}

	for i, v := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			iter := 0
			var w io.Writer = &bytes.Buffer{}
			if v.hasErr {
				w = &mockWriter{f: func() (int, error) {
					if iter > 0 {
						return 0, fmt.Errorf("some err")
					}
					iter++
					return 0, nil
				}}
			}
			err := PrintTree(w, v.m, "")
			if v.hasErr {
				assert.Error(t, err)
				return
			}
			b, ok := w.(*bytes.Buffer)
			if !assert.True(t, ok) {
				return
			}
			assert.Contains(t, v.expected, b.String())
		})
	}
}
