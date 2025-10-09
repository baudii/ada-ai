package utils

import (
	"os"
	"path/filepath"
	"testing"

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

func TestTrimJSON(t *testing.T) {
	t.Parallel()
	tests := []struct {
		json     string
		trimmed  []byte
		hasError bool
	}{
		{"{}", []byte("{}"), false},
		{"{", []byte("{"), true},
		{"}", []byte("}"), true},
		{"---{ \"k\": \"v\" }```-`", []byte("{ \"k\": \"v\" }"), false},
		{"abc", []byte("abc"), true},
	}

	for _, v := range tests {
		res, err := TrimJSON(v.json)
		assert.Equal(t, err != nil, v.hasError, "test: %v", v)
		assert.Equal(t, res, v.trimmed, "test: %v", v)
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

	for _, v := range tests {
		path := filepath.Join(t.TempDir(), "tmp.tmp")
		err := SaveJSONToFile(v.content, path)
		assert.Equal(t, err != nil, v.hasErr, "test: %v", v)
		if err == nil {
			got, readErr := os.ReadFile(path)
			assert.Equal(t, readErr == nil, v.fileExist, "test: %v", v)
			if readErr == nil {
				assert.Equal(t, []byte(v.result), got, "test: %v", v)
			}
		}
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

	for _, v := range tests {
		path := filepath.Join(t.TempDir(), "ttt.json")
		if v.hasFile {
			os.WriteFile(path, []byte(v.json), 0644)
		}
		res, err := ParseJSONFile[Valid](path)
		if v.hasErr {
			if v.hasFile {
				assert.Error(t, err, "test: %v", v)
			} else {
				assert.True(t, os.IsNotExist(err), "expected not exist error: test: %v", v)
			}
		} else {
			require.NoError(t, err, "test: %v", v)
			assert.Equal(t, *res, v.valid, "test: %v", v)
		}
	}
}

func TestMapToStruct(t *testing.T) {
	t.Parallel()
	type s struct {
		Val1 string         `json:"val1"`
		Val2 int            `json:"val2"`
		Val3 map[string]int `json:"val3"`
	}

	tests := []struct {
		m        map[string]any
		expected *s
		err      string
	}{
		{
			map[string]any{"val1": "s", "val2": 2, "val3": map[string]int{"r2": 1}},
			&s{Val1: "s", Val2: 2, Val3: map[string]int{"r2": 1}},
			"",
		},
		{
			map[string]any{"val1": make(chan int)},
			nil,
			"failed to marshal",
		},
		{
			map[string]any{"val2": "string"},
			nil,
			"failed to unmarshal into",
		},
	}

	for _, v := range tests {
		res, err := MapToStruct[s](v.m)
		if v.err == "" {
			assert.NoError(t, err, "test: %v", v)
		} else {
			assert.ErrorContains(t, err, v.err, "test: %v", v)
		}
		assert.Equal(t, v.expected, res, "test: %v", v)
	}
}
