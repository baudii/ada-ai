package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func ParseJsonFile[T any](filePath string) T {
	var cfg T
	path, err := filepath.Abs(filePath)
	if err == nil {
		fmt.Printf("Reading file %v\n", path)
	}

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
