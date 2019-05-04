package hash

import (
	"crypto/sha512"
	"fmt"
	"io"
	"os"
)

// Sum returns a SHA512 sum of a file.
func Sum(file string) (string, error) {
	var sum string
	f, err := os.Open(file)
	if err != nil {
		return sum, err
	}
	defer f.Close()
	h := sha512.New()
	if _, err := io.Copy(h, f); err != nil {
		return sum, err
	}
	sum = fmt.Sprintf("%x", h.Sum(nil))
	return sum, nil
}
