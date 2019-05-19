package executable

import "os"

// GetCwd returns the current working directory of the executable.
func GetCwd() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return dir, err
	}
	return dir, nil
}
