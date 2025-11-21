package fsutils

import "os"

func CopyFile(src, dst string) error {
	f, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.WriteFile(dst, f, 0o644); err != nil {
		return err
	}
	return nil
}
