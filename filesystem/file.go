package filesystem

import (
	"os"
	"path/filepath"
)

// FileExists checks if a file exists on the filesystem.
func FileExists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// Glob returns a list of all files within a directory.
func Glob(dir string) ([]string, error) {
	files := []string{}
	err := filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if f.IsDir() {
			return nil
		}
		files = append(files, path)
		return nil
	})
	return files, err
}
