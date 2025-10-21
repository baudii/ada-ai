package folder

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
)

// Mode defines the folder naming strategy.
type Mode int

type DirReader func(name string) ([]fs.DirEntry, error)

// Modes for folder naming strategy.
const (
	UseLatest Mode = iota
	CreateNew
)

// seqDir implements the DirProvider interface to provide
// sequentially numbered project folder names.
type seqDir struct {
	read DirReader
}

// NewSeqDir creates a new seqDir instance with the provided ReadDirFS.
func NewSeqDir(reader DirReader) *seqDir {
	return &seqDir{read: reader}
}

// ProjectFolder returns the folder name based on the existing directories
// in the base path. If createNew is true, it returns a new folder name
// by incrementing the highest existing integer-named folder. If false,
// it returns the highest existing integer-named folder or creates a new
// one if none exist.
func (d *seqDir) ProjectFolder(base string, mode Mode) (string, error) {
	dirEntries, err := d.read(base)
	if err != nil {
		if os.IsNotExist(err) {
			// If the directory does not exist, we can start with "0"
			return filepath.Join(base, "0"), nil
		}
		return "", fmt.Errorf("read dir %q: %w", base, err)
	}

	var lastIdx int
	switch mode {
	case CreateNew:
		lastIdx = newDir(dirEntries)
	case UseLatest:
		lastIdx = lastOrNew(dirEntries)
	default:
		return "", fmt.Errorf("unknown mode %v", mode)
	}
	return filepath.Join(base, strconv.Itoa(lastIdx)), nil
}

// newDir scans the provided directory entries for integer-named folders
// and returns the next new index by finding the highest integer N and
// returning N+1.
func newDir(entries []fs.DirEntry) int {
	max := -1
	for _, v := range entries {
		id, err := strconv.Atoi(v.Name())
		if err == nil && id > max {
			max = id
		}
	}

	return max + 1
}

// lastOrNew scans the provided directory entries for integer-named folders
// and returns the highest integer found. If no such folders exist, it
// returns the next new index by calling newDir.
func lastOrNew(entries []fs.DirEntry) int {
	max := -1
	for _, v := range entries {
		id, err := strconv.Atoi(v.Name())
		if err == nil && id > max && v.IsDir() {
			max = id
		}
	}

	if max == -1 {
		return newDir(entries)
	}

	return max
}
