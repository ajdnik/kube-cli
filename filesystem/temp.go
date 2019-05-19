package filesystem

import (
	"io/ioutil"
	"os"
)

// CreateTemp creates a temporaty file in the temporaty files
// directory and returns it's path.
func CreateTemp() (string, error) {
	var name string
	f, err := ioutil.TempFile(os.TempDir(), "kube-cli-")
	if err != nil {
		return name, err
	}
	name = f.Name()
	return name, nil
}
