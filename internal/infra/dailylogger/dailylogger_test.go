package dailylogger_test

import (
	"fmt"
	"io"
	"strconv"
	"testing"

	"github.com/baudii/ada-ai/internal/infra/dailylogger"
	"github.com/baudii/ada-ai/internal/infra/dailylogger/dailywriter"
	"github.com/stretchr/testify/require"
)

func TestDefaultSimpleLogger(t *testing.T) {
	t.Parallel()
	tests := []struct {
		cfg    *dailylogger.LogConfig
		opts   []dailywriter.Option
		hasErr bool
	}{
		{
			cfg: &dailylogger.LogConfig{Level: "debug", Timezone: "UTC"},
			opts: []dailywriter.Option{
				dailywriter.WithPreparer(func(path string) error { return nil }),
				dailywriter.WithOpener(func(path, prefix, date string) (io.WriteCloser, error) { return nil, nil }),
			},
			hasErr: false,
		},
		{
			cfg: &dailylogger.LogConfig{Level: "info", Timezone: "local"},
			opts: []dailywriter.Option{
				dailywriter.WithPreparer(func(path string) error { return nil }),
				dailywriter.WithOpener(func(path, prefix, date string) (io.WriteCloser, error) { return nil, nil }),
			},
			hasErr: false,
		},
		{
			cfg: &dailylogger.LogConfig{Level: "warn", Timezone: ".."},
			opts: []dailywriter.Option{
				dailywriter.WithPreparer(func(path string) error { return nil }),
				dailywriter.WithOpener(func(path, prefix, date string) (io.WriteCloser, error) { return nil, nil }),
			},
			hasErr: false,
		},
		{
			cfg: &dailylogger.LogConfig{Level: "error", Timezone: "America/New_York"},
			opts: []dailywriter.Option{
				dailywriter.WithPreparer(func(path string) error { return nil }),
				dailywriter.WithOpener(func(path, prefix, date string) (io.WriteCloser, error) { return nil, nil }),
			},
			hasErr: false,
		},
		{
			cfg: &dailylogger.LogConfig{},
			opts: []dailywriter.Option{
				dailywriter.WithPreparer(func(path string) error { return fmt.Errorf("mock") }),
			},
			hasErr: true,
		},
	}
	for i, v := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			logger, err := dailylogger.DefaultDailyLogger(v.cfg, v.opts...)
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
