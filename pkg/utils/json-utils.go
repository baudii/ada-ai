package utils

import (
	"encoding/json"
	"os"
)

func SaveJsonToFile(content any, filePath string) error {
	_, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(content, "", "\t")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0)
}

func ParseJsonFile[T any](filePath string) (*T, error) {
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
