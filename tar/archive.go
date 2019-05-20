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

// Read file contents and add it to the archive.
func addFile(tw *tar.Writer, path string, base *string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	if stat, err := file.Stat(); err == nil {
		// Create archive header for file
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
		// Write the header to the archive
		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		// Copy file contents to the archive
		if _, err := io.Copy(tw, file); err != nil {
			return err
		}
	}
	return nil
}
