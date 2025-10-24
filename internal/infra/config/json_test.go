package config_test

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/baudii/ada-ai/internal/infra/config"
	"github.com/baudii/ada-ai/internal/infra/folders"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Bad struct {
	C chan string
}
type Valid struct {
	Name any    `json:"name"`
	Age  int    `json:"age"`
	C    func() `json:"-"`
}

type Nested struct {
	M Valid `json:"m"`
}

func TestLoadWithLocal(t *testing.T) {
	t.Parallel()
	type mockCfg struct {
		A string         `json:"a"`
		B int            `json:"b"`
		C map[string]int `json:"c"`
	}

	tests := []struct {
		name         string
		expected     mockCfg
		baseContent  string
		localContent string
		err          string
	}{
		{
			name: "success: merge base and local json",
			expected: mockCfg{
				A: "world",
				B: 1,
				C: map[string]int{"d": 100},
			},
			baseContent:  "{\"a\":\"hello\", \"b\":1,\"c\": {\"d\":100}}",
			localContent: "{\"a\":\"world\"}",
		},
		{
			name: "success: merge base and local json with missing fields",
			expected: mockCfg{
				A: "hello",
				B: 1,
				C: map[string]int{"d": 100},
			},
			baseContent:  "{\"a\":\"hello\", \"b\":1}",
			localContent: "{\"c\": {\"d\":100}}",
		},
		{
			name: "fail: file does not exist",
			err:  "open",
		},
		{
			name:         "fail: invalid main json",
			baseContent:  "{",
			localContent: "{}",
			err:          "unmarshal",
		},
		{
			name:         "fail: invalid local json",
			baseContent:  "{}",
			localContent: "}",
			err:          "unmarshal",
		},
		{
			name:        "fail: local not exist",
			baseContent: `{"A":"value"}`,
			expected:    mockCfg{A: "value"},
		},
	}

	for _, v := range tests {
		t.Run(v.name, func(t *testing.T) {
			t.Parallel()
			path := t.TempDir()
			path = filepath.Join(path, "tmp.json")
			if v.baseContent != "" {
				err := os.WriteFile(path, []byte(v.baseContent), 0644)
				require.NoError(t, err)
			}
			if v.localContent != "" {
				localPath := folders.InsertFsuffix(path, ".local")
				err := os.WriteFile(localPath, []byte(v.localContent), 0644)
				require.NoError(t, err)
			}

			res, err := config.LoadWithLocal[mockCfg](path)
			if v.err != "" {
				assert.ErrorContains(t, err, v.err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, v.expected, res)
			}
		})
	}
}

func TestParseFileToMap(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		path        string
		expected    map[string]any
		jsonContent string
		err         string
	}{
		{
			name: "fail: file does not exist",
			path: t.TempDir(),
			err:  "open",
		},
		{
			name:        "fail: invalid json",
			path:        t.TempDir(),
			jsonContent: "{",
			err:         "unmarshal",
		},
		{
			name:        "success: valid json",
			path:        t.TempDir(),
			expected:    map[string]any{"a": "hello", "b": "world", "c": map[string]any{"d": float64(100)}},
			jsonContent: "{\"a\":\"hello\", \"b\":\"world\",\"c\": {\"d\":100}}",
			err:         "",
		},
	}

	for _, v := range tests {
		t.Run(v.name, func(t *testing.T) {
			t.Parallel()
			v.path = filepath.Join(v.path, "tmp.json")
			if v.jsonContent != "" {
				err := os.WriteFile(v.path, []byte(v.jsonContent), 0644)
				require.NoError(t, err)
			}

			res, err := config.Load[map[string]any](v.path)
			if v.err != "" {
				assert.ErrorContains(t, err, v.err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, v.expected, res)
		})
	}
}

func TestSaveJSONToFile(t *testing.T) {
	t.Parallel()
	tests := []struct {
		content   any
		result    string
		fileExist bool
		hasErr    bool
	}{
		{&Nested{Valid{"a", 12, func() {}}}, "{\n\t\"m\": {\n\t\t\"name\": \"a\",\n\t\t\"age\": 12\n\t}\n}", true, false},
		{func() {}, "", false, true},
		{new(Bad), "", false, true},
		{map[int]string{1: "one"}, "{\n\t\"1\": \"one\"\n}", true, false},
		{"{", "\"{\"", true, false},
	}

	for i, v := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "tmp.tmp")
			err := config.Save(v.content, path)
			assert.Equal(t, err != nil, v.hasErr)
			if err == nil {
				got, readErr := os.ReadFile(path)
				assert.Equal(t, readErr == nil, v.fileExist)
				if readErr == nil {
					assert.Equal(t, []byte(v.result), got)
				}
			}
		})
	}
}

func TestParseJSONFile(t *testing.T) {
	t.Parallel()
	tests := []struct {
		json    string
		valid   Valid
		hasFile bool
		hasErr  bool
	}{
		{"{\"name\": \"alice\", \"age\":30}", Valid{"alice", 30, nil}, true, false},
		{"{\"name\": \"alice\" }", Valid{"alice", 0, nil}, true, false},
		{"{}", Valid{}, true, false},
		{"", Valid{}, true, true},
		{"{\"name\": \"alice\" }", Valid{"alice", 0, nil}, false, true},
	}

	for i, v := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "ttt.json")
			if v.hasFile {
				err := os.WriteFile(path, []byte(v.json), 0644)
				require.NoError(t, err)
			}
			res, err := config.Load[Valid](path)
			if v.hasErr {
				if v.hasFile {
					assert.Error(t, err)
				} else {
					assert.True(t, errors.Is(err, fs.ErrNotExist), "expected not exist error: test: %v", v)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, res, v.valid)
			}
		})
	}
}

func TestMapToStruct(t *testing.T) {
	t.Parallel()
	type mockType struct {
		Val1 string         `json:"val1"`
		Val2 int            `json:"val2"`
		Val3 map[string]int `json:"val3"`
	}

	tests := []struct {
		name     string
		m        map[string]any
		expected mockType
		err      string
	}{
		{
			name:     "valid",
			m:        map[string]any{"val1": "s", "val2": 2, "val3": map[string]int{"r2": 1}},
			expected: mockType{Val1: "s", Val2: 2, Val3: map[string]int{"r2": 1}},
			err:      "",
		},
		{
			name: "fail: marshal",
			m:    map[string]any{"val1": make(chan int)},
			err:  "marshal",
		},
		{
			name: "fail: unmarshal",
			m:    map[string]any{"val2": "string"},
			err:  "unmarshal into",
		},
	}

	for _, v := range tests {
		t.Run(v.name, func(t *testing.T) {
			res, err := config.UnmarshalMap[mockType](v.m)
			if v.err == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, v.err)
			}
			assert.Equal(t, v.expected, res)
		})
	}
}
