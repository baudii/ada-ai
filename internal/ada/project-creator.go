package ada

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/baudii/ada-ai/internal"
)

var projDescrFile string
var projRoot string

func projectExist() bool {
	path := filepath.Join(cfg.ProjRoot, cfg.UserName, cfg.ProjName, "project-structure.yml")
	exist := internal.PathExist(path)
	if exist {
		projDescrFile = path
		projRoot = filepath.Dir(path)
	}
	return exist
}

func parseStructure() error {
	file, err := os.Open(projDescrFile)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(file)
	prevIndent := 0
	curPath := ""
	folder := ""
	for scanner.Scan() {
		line := scanner.Text()
		indent, err := getIndent(line, 2)
		if err != nil || indent-prevIndent > 1 {
			return fmt.Errorf("incorrect indent provided")
		}

		for indent != prevIndent {
			if indent < prevIndent {
				curPath = filepath.Dir(curPath)
				prevIndent--
			} else if indent > prevIndent {
				curPath = filepath.Join(curPath, folder)
				prevIndent++
				if prevIndent != indent {
					return fmt.Errorf("incorrect indent provided")
				}
				break
			}
		}

		line = strings.TrimSpace(line)
		f := filepath.Join(projRoot, curPath, line)
		if line[len(line)-1] == '/' {
			folder = filepath.Dir(line)
			os.MkdirAll(f, 0o755)
		} else {
			os.Create(f)
		}

		fmt.Println(f)
	}

	return nil
}

func getIndent(line string, size int) (int, error) {
	indent := 0
	flag := 0
	for _, c := range line {
		if unicode.IsSpace(c) {
			if flag == size-1 {
				indent++
			}
			flag = (flag + 1) % size
		} else {
			break
		}
	}
	if flag != 0 {
		return 0, fmt.Errorf("invalid indentation")
	}
	return indent, nil
}

func createDirectory() error {
	path := filepath.Join(cfg.ProjRoot, cfg.UserName, cfg.ProjName)
	_, err := os.Stat(path)
	if err == nil {
		return fmt.Errorf("project already exists with the same name %v", path)
	}

	if !os.IsNotExist(err) {
		return err
	}

	err = os.MkdirAll(path, 0o755)
	if err != nil {
		return err
	}

	projRoot = path
	return nil
}

func saveProjectStructure(yml string) error {
	projDescrFile = filepath.Join(projRoot, "project-structure.yml")
	return os.WriteFile(projDescrFile, []byte(yml), 0644)
}
