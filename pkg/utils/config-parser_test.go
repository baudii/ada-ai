package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseJSONFileToMap(t *testing.T) {
	t.Parallel()
	tests := []struct {
		path        string
		expected    map[string]any
		jsonContent string
		hasErr      bool
	}{
		{t.TempDir(), nil, "", true},
		{t.TempDir(), nil, "{", true},
		{t.TempDir(), map[string]any{
			"a": "hello",
			"b": "world",
			"c": map[string]any{"d": float64(100)},
		}, "{\"a\":\"hello\", \"b\":\"world\",\"c\": {\"d\":100}}", false},
	}

	for i, v := range tests {
		v.path = filepath.Join(v.path, "tmp.json")
		if v.jsonContent != "" {
			err := os.WriteFile(v.path, []byte(v.jsonContent), 0644)
			if !assert.NoError(t, err) {
				continue
			}
		}

		res, err := ParseJSONFileToMap(v.path)
		assert.True(t, (err != nil) == v.hasErr, "test: %v", i)
		assert.Equal(t, v.expected, res, "test: %v", i)
	}
}

func TestParseJSONConfigWithLocal(t *testing.T) {
	t.Parallel()
	type config struct {
		A string         `json:"a"`
		B int            `json:"b"`
		C map[string]int `json:"c"`
	}

	tests := []struct {
		path             string
		expected         *config
		jsonContent      string
		localJsonContent string
		hasErr           bool
	}{
		{t.TempDir(), nil, "", "", true},
		{t.TempDir(), nil, "{", "{}", true},
		{t.TempDir(), nil, "{}", "}", true},
		{t.TempDir(), nil, "{\"b\":\"fail\"}", "{}", true},
		{
			t.TempDir(),
			&config{
				A: "world",
				B: 1,
				C: map[string]int{"d": 100},
			},
			"{\"a\":\"hello\", \"b\":1,\"c\": {\"d\":100}}",
			"{\"a\":\"world\"}",
			false,
		},
		{
			t.TempDir(),
			&config{
				A: "hello",
				B: 1,
				C: map[string]int{"d": 100},
			},
			"{\"a\":\"hello\", \"b\":1}",
			"{\"c\": {\"d\":100}}",
			false,
		},
	}

	for i, v := range tests {
		v.path = filepath.Join(v.path, "tmp.json")
		if v.jsonContent != "" {
			localPath := InsertFsuffix(v.path, ".local")
			err := os.WriteFile(v.path, []byte(v.jsonContent), 0644)
			if !assert.NoError(t, err) {
				continue
			}
			err = os.WriteFile(localPath, []byte(v.localJsonContent), 0644)
			if !assert.NoError(t, err) {
				continue
			}
		}

		res, err := ParseJSONConfigWithLocal[config](v.path)
		assert.True(t, (err != nil) == v.hasErr, "test: %v", i)
		assert.Equal(t, v.expected, res, "test: %v", i)
	}
}
