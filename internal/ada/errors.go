package ada

import "fmt"

type projExist struct {
	Path string
}

func (e *projExist) Error() string {
	return fmt.Sprintf("project already exists at %v", e.Path)
}
