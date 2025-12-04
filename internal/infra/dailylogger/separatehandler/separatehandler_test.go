package separatehandler_test

import (
	"context"
	"log/slog"
	"strings"
	"testing"

	"github.com/baudii/ada-ai/internal/infra/dailylogger/separatehandler"
	"github.com/stretchr/testify/assert"
)

// MockBrokenWriter is an io.Writer that can be configured to return an error.
type MockBrokenWriter struct {
	ShouldError   bool
	ErrorToReturn error
}

// Write implements the io.Writer interface for MockBrokenWriter.
func (m *MockBrokenWriter) Write(p []byte) (n int, err error) {
	if m.ShouldError {
		return 0, m.ErrorToReturn
	}

	return len(p), nil
}

func TestWithLevelWriters(t *testing.T) {
	t.Parallel()
	infoWriter, errorWriter := &strings.Builder{}, &strings.Builder{}

	tests := []struct {
		name     string
		writers  []any
		messages []string
		levels   []slog.Level
		expected [][]string
	}{
		{
			name: "info and error writers",
			writers: []any{
				slog.LevelInfo, infoWriter, slog.LevelError, errorWriter,
			},
			messages: []string{"info message", "error message"},
			levels:   []slog.Level{slog.LevelInfo, slog.LevelError},
			expected: [][]string{{"info message", "error message"}, {"error message"}},
		},
		{
			name: "only error writer",
			writers: []any{
				slog.LevelError, errorWriter,
			},
			messages: []string{"info message", "error message"},
			levels:   []slog.Level{slog.LevelInfo, slog.LevelError},
			expected: [][]string{{}, {"error message"}},
		},
	}

	for _, v := range tests {
		t.Run(v.name, func(t *testing.T) {
			infoWriter.Reset()
			errorWriter.Reset()

			sh := separatehandler.New(separatehandler.WithLevelWriters(v.writers...))
			h := sh.WithGroup("no op")
			s := slog.New(h)
			for i, msg := range v.messages {
				s.Log(context.Background(), v.levels[i], msg)
			}

			splInfo := strings.Split(infoWriter.String(), "\n")
			splError := strings.Split(errorWriter.String(), "\n")
			assert.Equal(t, len(v.expected[0]), len(splInfo)-1)
			assert.Equal(t, len(v.expected[1]), len(splError)-1)
			for _, expectedMsg := range v.expected[0] {
				assert.Contains(t, infoWriter.String(), expectedMsg)
			}
			for _, expectedMsg := range v.expected[1] {
				assert.Contains(t, errorWriter.String(), expectedMsg)
			}
		})
	}
}
