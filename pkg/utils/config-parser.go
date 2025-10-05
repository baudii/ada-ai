package utils

import (
	"encoding/json"
	"fmt"
	"os"
)

// Parses a file at basePath, then merges it with file at the same root folder but named
// filename.local.ext where filename is a filename of basePath and ext is an extensions of basePath.
// Then tries to parse local configuration and performs a merge by calling MergeMap(). All values of
// local configuration file will overwrite the values from base configuration.
//
// If local file doesn't exist only basePath will be parsed. If the local file is invalid JSON it will
// return an error to ensure correct behavior.
//
// json.Marshal on a map parsed from JSON should never fail.
// This check only protects against unexpected values introduced by MergeMap or code changes.
func ParseJSONConfigWithLocal[T any](basePath string) (*T, error) {
	basecfg, err := ParseJSONFileToMap(basePath)
	if err != nil {
		return nil, fmt.Errorf("read base config: %w", err)
	}

	localPath := InsertFsuffix(basePath, ".local")
	if localcfg, err := ParseJSONFileToMap(localPath); err == nil {
		MergeMap(basecfg, localcfg)
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to read '%v' error: %w", localPath, err)
	}

	return MapToStruct[T](basecfg)
}

// Deserializes a JSON file into a map[string]any object.
func ParseJSONFileToMap(path string) (map[string]any, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var m map[string]any
	if err := json.NewDecoder(f).Decode(&m); err != nil {
		return nil, fmt.Errorf("unmarshal %s: %w", path, err)
	}

	return m, nil
}
