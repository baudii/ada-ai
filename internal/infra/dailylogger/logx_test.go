package dailylogger_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/baudii/ada-ai/internal/infra/dailylogger"
	"github.com/stretchr/testify/require"
)

func TestLoadLogConfigOrDefault(t *testing.T) {
	t.Parallel()

	tests := []struct {
		cfgPath  string
		expected dailylogger.LogConfig
		hasErr   bool
	}{
		{
			hasErr:   false,
			cfgPath:  "testdata/valid_config.json",
			expected: dailylogger.LogConfig{Timezone: "UTC", Path: "logs", Prefix: "prefix"},
		},
		{
			hasErr:   true,
			cfgPath:  "testdata/invalid_config.json",
			expected: dailylogger.DefaultCfg,
		},
	}

	for i, v := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
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
			got := dailylogger.LoadLogConfigOrDefault(v.cfgPath)
			require.Equal(t, v.expected, got)
		})
	}
}
