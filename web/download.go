package web

import (
	"io"
	"net/http"
	"os"
)

// DownloadFile downloads a file from a given url and saves it in the
// local filesystem.
func DownloadFile(file, url string) (int64, error) {
	// Create the file, but give it a tmp file extension, this means we won't overwrite a
	// file until it's downloaded, but we'll remove the tmp extension once downloaded.
	out, err := os.Create(file + ".tmp")
	if err != nil {
		return -1, err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return -1, err
	}

	err = os.Rename(file+".tmp", file)
	if err != nil {
		return -1, err
	}

	return resp.ContentLength, nil
}
