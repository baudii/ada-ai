package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func ParseJSONConfigWithLocal[T any](basePath string) (*T, error) {
	baseMap, err := ParseJSONFileToMap(basePath)
	if err != nil {
		return nil, fmt.Errorf("read base config: %w", err)
	}

	ext := filepath.Ext(basePath)
	name := strings.TrimSuffix(basePath, ext)
	localPath := name + ".local" + ext
	if st, err := os.Stat(localPath); err == nil && !st.IsDir() {
		localMap, err := ParseJSONFileToMap(localPath)
		if err != nil {
			return nil, fmt.Errorf("read local config: %w", err)
		}
		DeepMerge(baseMap, localMap)
	}

	mergedBytes, err := json.Marshal(baseMap)
	if err != nil {
		return nil, fmt.Errorf("marshal merged config: %w", err)
	}

	var cfg T
	if err := json.Unmarshal(mergedBytes, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal merged into T: %w", err)
	}

	return &cfg, nil
}
