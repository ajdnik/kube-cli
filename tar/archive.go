package tar

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
)

// Archive compresses a folder to .tar.gz archive.
func Archive(files []string, arch string, base *string) error {
	f, err := os.Create(arch)
	if err != nil {
		return err
	}
	defer f.Close()
	gw := gzip.NewWriter(f)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()
	for i := range files {
		if err := addFile(tw, files[i], base); err != nil {
			return err
		}
	}
	return nil
}

func addFile(tw *tar.Writer, path string, base *string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	if stat, err := file.Stat(); err == nil {
		// now lets create the header as needed for this file within the tarball
		header := new(tar.Header)
		header.Name = path
		if base != nil {
			rel, err := filepath.Rel(*base, path)
			if err != nil {
				return err
			}
			header.Name = rel
		}
		header.Size = stat.Size()
		header.Mode = int64(stat.Mode())
		header.ModTime = stat.ModTime()
		// write the header to the tarball archive
		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		// copy the file data to the tarball
		if _, err := io.Copy(tw, file); err != nil {
			return err
		}
	}
	return nil
}
