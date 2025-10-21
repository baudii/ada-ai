package dailywriter

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockWriteCloser struct {
	io.Writer
}

func (mockWriteCloser) Close() error { return nil }

func TestNew(t *testing.T) {
	t.Parallel()
	tests := []struct {
		opts     []Option
		expected *dailyWriter
		err      string
	}{
		{
			opts: []Option{
				WithPrefix("a"),
				WithLocation(time.Local),
				WithDater(func(l *time.Location) string { return "2022-12-10" }),
				WithPreparer(func(path string) error { return nil }),
				WithOpener(func(path, prefix, date string) (io.WriteCloser, error) { return mockWriteCloser{io.Discard}, nil }),
			},
			expected: &dailyWriter{prefix: "a", loc: time.Local},
			err:      "",
		},
		{
			opts: []Option{
				WithPreparer(func(path string) error { return fmt.Errorf("prepare mock") }),
			},
			expected: nil,
			err:      "new dailywriter: rotate prepare: prepare mock",
		},
		{
			opts: []Option{
				WithPreparer(func(path string) error { return nil }),
				WithOpener(func(path, prefix, date string) (io.WriteCloser, error) { return nil, fmt.Errorf("open mock") }),
			},
			expected: nil,
			err:      "new dailywriter: rotate open: open mock",
		},
	}

	for i, v := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			d := t.TempDir()
			dw, err := New(d, v.opts...)
			if v.err == "" {
				assert.NoError(t, err)
				assert.NotNil(t, dw)
				assert.Equal(t, v.expected.prefix, dw.prefix)
				assert.Equal(t, v.expected.loc, dw.loc)
				assert.NotNil(t, dw.wc)
			} else {
				assert.ErrorContains(t, err, v.err)
			}
		})
	}
}

func TestWrite(t *testing.T) {
	t.Parallel()
	tests := []struct {
		opts     []Option
		input    string
		expected string
		swappath string
		err      string
	}{
		{
			opts: []Option{
				WithPreparer(func(path string) error { return nil }),
				WithOpener(func(path, prefix, date string) (io.WriteCloser, error) { return mockWriteCloser{&bytes.Buffer{}}, nil }),
			},
			input:    "test",
			expected: "test",
			swappath: "",
			err:      "",
		},
		{
			opts: []Option{
				WithPreparer(func(path string) error {
					if path == "fail" {
						return fmt.Errorf("mock")
					}
					return nil
				}),
				WithOpener(func(path, prefix, date string) (io.WriteCloser, error) { return mockWriteCloser{&bytes.Buffer{}}, nil }),
			},
			input:    "test",
			expected: "",
			swappath: "fail",
			err:      "rotate prepare: mock",
		},
	}

	for i, v := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			d := t.TempDir()
			dw, err := New(d, v.opts...)
			require.NoError(t, err)
			if v.swappath != "" {
				dw.path = v.swappath
				dw.curDate = "2022-01-01" // force rotation
			}
			n, err := dw.Write([]byte(v.input))
			if v.err == "" {
				if assert.NoError(t, err) {
					assert.Equal(t, len(v.input), n)
					buf := dw.wc.(mockWriteCloser).Writer.(*bytes.Buffer)
					assert.Equal(t, v.expected, buf.String())
				}
			} else {
				assert.ErrorContains(t, err, v.err)
			}
		})
	}
}

func TestWrite_DefaultPrepareOpener(t *testing.T) {
	t.Parallel()
	d := t.TempDir()
	dw, err := New(
		d,
		WithPrefix("m"),
		WithDater(func(loc *time.Location) string { return "2020-01-02" }),
	)
	require.NoError(t, err)
	require.NotNil(t, dw)
	n, err := dw.Write([]byte("test"))
	require.NoError(t, err)
	require.Equal(t, 4, n)
	require.NotNil(t, dw.wc)
	f, err := os.ReadFile(filepath.Join(d, "m_2020-01-02.log"))
	require.NoError(t, err)
	require.Equal(t, "test", string(f))
	_ = dw.wc.Close() // Close the writer to avoid resource leaks
}
