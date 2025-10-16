package common

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/baudii/ada-ai/pkg/dilog"
	"github.com/stretchr/testify/require"
)

func TestLoadLogConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		cfgPath  string
		expected *dilog.LogConfig
		hasErr   bool
	}{
		{
			hasErr:   false,
			cfgPath:  "testdata/valid_config.json",
			expected: &dilog.LogConfig{Timezone: "UTC", Path: "logs", Prefix: "prefix"},
		},
		{
			hasErr:   true,
			cfgPath:  "testdata/invalid_config.json",
			expected: &defaultCfg,
		},
	}

	for _, v := range tests {
		t.Run(v.cfgPath, func(t *testing.T) {
			if !v.hasErr {
				d := t.TempDir()
				v.cfgPath = filepath.Join(d, v.cfgPath)
				err := os.MkdirAll(filepath.Dir(v.cfgPath), 0o755)
				require.NoError(t, err)
				f, err := os.Create(v.cfgPath)
				require.NoError(t, err)
				j, err := json.Marshal(v.expected)
				require.NoError(t, err)
				_, err = f.Write(j)
				require.NoError(t, err)
				err = f.Close()
				require.NoError(t, err)
			}
			got := LoadLogConfig(v.cfgPath)
			require.Equal(t, v.expected, got)
		})
	}
}

func TestDefaultSimpleLogger(t *testing.T) {
	t.Parallel()
	tests := []struct {
		cfg    *dilog.LogConfig
		opts   []dilog.Option
		hasErr bool
	}{
		{
			cfg: &dilog.LogConfig{Level: "debug", Timezone: "UTC"},
			opts: []dilog.Option{
				dilog.WithPreparer(func(path string) error { return nil }),
				dilog.WithOpener(func(path, prefix, date string) (io.WriteCloser, error) { return nil, nil }),
			},
			hasErr: false,
		},
		{
			cfg: &dilog.LogConfig{Level: "info", Timezone: "local"},
			opts: []dilog.Option{
				dilog.WithPreparer(func(path string) error { return nil }),
				dilog.WithOpener(func(path, prefix, date string) (io.WriteCloser, error) { return nil, nil }),
			},
			hasErr: false,
		},
		{
			cfg: &dilog.LogConfig{Level: "warn", Timezone: ".."},
			opts: []dilog.Option{
				dilog.WithPreparer(func(path string) error { return nil }),
				dilog.WithOpener(func(path, prefix, date string) (io.WriteCloser, error) { return nil, nil }),
			},
			hasErr: false,
		},
		{
			cfg: &dilog.LogConfig{Level: "error", Timezone: "America/New_York"},
			opts: []dilog.Option{
				dilog.WithPreparer(func(path string) error { return nil }),
				dilog.WithOpener(func(path, prefix, date string) (io.WriteCloser, error) { return nil, nil }),
			},
			hasErr: false,
		},
		{
			cfg: &dilog.LogConfig{},
			opts: []dilog.Option{
				dilog.WithPreparer(func(path string) error { return fmt.Errorf("mock") }),
			},
			hasErr: true,
		},
	}
	for i, v := range tests {
		t.Run(string(rune(i)), func(t *testing.T) {
			logger, err := DefaultSimpleLogger(v.cfg, v.opts...)
			if v.hasErr {
				require.Error(t, err)
				require.Nil(t, logger)
			} else {
				require.NoError(t, err)
				require.NotNil(t, logger)
			}
		})
	}
}
