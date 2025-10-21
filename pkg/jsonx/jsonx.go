package jsonx

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/baudii/ada-ai/pkg/maps"
	"github.com/baudii/ada-ai/pkg/pathx"
)

// LoadWithLocal parses a file at basePath, then merges it with file at the same root folder but named
// filename.local.ext where filename is a filename of basePath and ext is an extensions of basePath.
// Then tries to parse local configuration and performs a merge by calling MergeMap(). All values of
// local configuration file will overwrite the values from base configuration.
//
// If local file doesn't exist only basePath will be parsed. If the local file is invalid JSON it will
// return an error to ensure correct behavior.
//
// json.Marshal on a map parsed from JSON should never fail.
// This check only protects against unexpected values introduced by MergeMap or code changes.
func LoadWithLocal[T any](path string) (T, error) {
	var zero T
	m, err := LoadWithLocalToMap(path)
	if err != nil {
		return zero, err
	}

	return UnmarshalMap[T](m)
}

// LoadWithLocalToMap parses a file at basePath, then merges it with file at the same root folder but named
// filename.local.ext where filename is a filename of basePath and ext is an extensions of basePath.
// Then tries to parse local configuration and performs a merge by calling MergeMap(). All values of
// local configuration file will overwrite the values from base configuration.
func LoadWithLocalToMap(path string) (map[string]any, error) {
	basecfg, err := Load[map[string]any](path)
	if err != nil {
		return nil, fmt.Errorf("read base config: %w", err)
	}

	localPath := pathx.InsertFsuffix(path, ".local")
	localcfg, err := Load[map[string]any](localPath)
	switch {
	case err == nil:
		maps.Merge(basecfg, localcfg)
		return basecfg, nil
	case errors.Is(err, fs.ErrNotExist):
		return basecfg, nil
	default:
		return nil, fmt.Errorf("read local config: %w", err)
	}
}

// Load reads a JSON file from the specified filePath and unmarshals
// its content into a struct of type T. It returns a pointer to the struct
// and any error encountered during reading or unmarshalling.
func Load[T any](filePath string) (T, error) {
	var res T
	file, err := os.ReadFile(filePath)
	if err != nil {
		return res, fmt.Errorf("read file: %w", err)
	}

	err = json.Unmarshal(file, &res)
	if err != nil {
		return res, fmt.Errorf("unmarshal %q: %w", filePath, err)
	}

	return res, nil
}

// Save marshals the given content into a pretty-printed JSON format
// and writes it to the specified file path. It returns an error if the
// marshalling or file writing fails.
func Save(content any, filePath string) error {
	data, err := json.MarshalIndent(content, "", "\t")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0o644)
}

// Trim attempts to extract a valid JSON object from a given string.
// It looks for the first occurrence of '{' and the last occurrence of '}'.
// If both are found, it returns the substring between these indices (inclusive).
// If either character is not found, it returns an error indicating that
// the input is not a valid JSON.
func Trim(data string) ([]byte, error) {
	start := strings.IndexByte(data, '{')
	end := strings.LastIndexByte(data, '}')
	if start == -1 || end == -1 {
		return []byte(data), fmt.Errorf("not a valid json: failed to find '{' or '}'")
	}
	return []byte(data)[start : end+1], nil
}

// UnmarshalMap converts given map m into a struct of type T. Uses json marshalling
// and unmarshalling under the hood.
func UnmarshalMap[T any](m map[string]any) (T, error) {
	var res T
	marshalled, err := json.Marshal(m)
	if err != nil {
		return res, fmt.Errorf("marshal: %w", err)
	}

	if err := json.Unmarshal(marshalled, &res); err != nil {
		return res, fmt.Errorf("unmarshal into %T: %w", res, err)
	}

	return res, nil
}
