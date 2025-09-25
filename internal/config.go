package internal

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const Artifacts string = "artifacts"

func GetAbsolutePath(relativePath string) string {
	exe, err := os.Executable()
	if err != nil {
		panic(err)
	}

	base := filepath.Dir(exe)
	return filepath.Join(base, relativePath)
}

func PathExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

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

func ToAnySlice(ss []string) []any {
	res := make([]any, len(ss))
	for i, v := range ss {
		res[i] = v
	}
	return res
}
