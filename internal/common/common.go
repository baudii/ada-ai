package common

import "path/filepath"

// Define the path to the named folders. Used for centrilized access to the files that contain data.
var (
	DataPath      = "data"
	ConfigPath    = filepath.Join(DataPath, "configuration")
	DebuggingPath = filepath.Join(DataPath, "debugging")
	PromptsPath   = filepath.Join(DataPath, "prompts")
)
