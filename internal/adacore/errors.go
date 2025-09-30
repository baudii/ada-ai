package adacore

import "fmt"

type ProjExist struct {
	Path string
}

func (e *ProjExist) Error() string {
	return fmt.Sprintf("project already exists at %v", e.Path)
}
