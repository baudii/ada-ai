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
		return []byte(data), fmt.Errorf("not a valid json: failed to find '{' or '}'")
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
	var res T
	file, err := os.ReadFile(filePath)
	if err != nil {
		return &res, err
	}

	err = json.Unmarshal(file, &res)
	if err != nil {
		return &res, err
	}

	return &res, nil
}

// Converts given map m into a struct of type T. Uses json marshalling
// and unmarshalling under the hood.
func MapToStruct[T any](m map[string]any) (*T, error) {
	marshalled, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal: %w", err)
	}

	var cfg T
	if err := json.Unmarshal(marshalled, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal into %T: %w", cfg, err)
	}

	return &cfg, nil
}
