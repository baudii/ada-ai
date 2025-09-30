package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func TrimJSON(data string) ([]byte, error) {
	start := strings.IndexByte(data, '{')
	end := strings.LastIndexByte(data, '}')
	if start == -1 || end == -1 {
		return nil, fmt.Errorf("not a valid json")
	}
	return []byte(data)[start : end+1], nil
}

func SaveJSONToFile(content any, filePath string) error {
	data, err := json.MarshalIndent(content, "", "\t")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0o644)
}

func ParseJSONFile[T any](filePath string) (*T, error) {
	var cfg T
	file, err := os.ReadFile(filePath)
	if err != nil {
		return &cfg, err
	}

	err = json.Unmarshal(file, &cfg)
	if err != nil {
		return &cfg, err
	}

	return &cfg, nil
}

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
