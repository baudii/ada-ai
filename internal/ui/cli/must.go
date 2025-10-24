package cli

import "fmt"

// Must is a helper function that panics if the provided error is not nil.
// It returns the value of type T if there is no error.
func Must[T any](v T, err error) T {
	if err != nil {
		panic(fmt.Sprintf("Must failed: %v", err))
	}
	return v
}

// MustErr is a helper function that panics if the provided error is not nil.
func MustErr(err error) {
	if err != nil {
		panic(fmt.Sprintf("MustErr failed: %v", err))
	}
}
