package tar

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
)

// Unarchive extracts files and folders from a .tar.gz archive.
func Unarchive(file, path string) error {
	ar, err := os.Open(file)
	if err != nil {
		return err
	}
	defer ar.Close()
	gzf, err := gzip.NewReader(ar)
	if err != nil {
		return err
	}
	re := tar.NewReader(gzf)
	for true {
		header, err := re.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		switch header.Typeflag {
		case tar.TypeDir:
			err = os.Mkdir(filepath.Join(path, header.Name), 0755)
			if err != nil {
				return err
			}
		case tar.TypeReg:
			f, err := os.Create(filepath.Join(path, header.Name))
			if err != nil {
				return err
			}
			defer f.Close()
			if _, err := io.Copy(f, re); err != nil {
				return err
			}
		}
	}
	return nil
}
