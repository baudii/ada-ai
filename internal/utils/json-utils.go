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

func ParseJsonFile[T any](filePath string) T {
	var cfg T
	file, err := os.ReadFile(filePath)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(file, &cfg)
	if err != nil {
		panic(err)
	}

	return cfg
}
