package simplehandler_test

import (
	"bytes"
	"log/slog"
	"regexp"
	"testing"
	"time"

	"github.com/baudii/ada-ai/internal/infra/dailylogger/simplehandler"
	"github.com/stretchr/testify/assert"
)

func TestSimpleHandler(t *testing.T) {
	t.Parallel()
	b := &bytes.Buffer{}
	expected := regexp.MustCompile(`^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}.*key1=value1 key2=value2.*\n?$`)

	// Create a SimpleHandler with a buffer as the writer
	sh := simplehandler.New(b, simplehandler.WithLevel(slog.LevelDebug), simplehandler.WithLocation(time.UTC))
	h := sh.WithGroup("no op").WithAttrs([]slog.Attr{{Key: "key1", Value: slog.StringValue("value1")}})
	s := slog.New(h)
	s.Info("test message", slog.String("key2", "value2"))
	assert.Regexp(t, expected, b.String())
}
