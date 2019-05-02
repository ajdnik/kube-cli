// Copyright © 2019 Rok Ajdnik and contributors
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha512"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/ajdnik/kube-cli/version"
	humanize "github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update the command line tool",
	Long: `Update the command line tool by pulling the latest
version from the web. Make sure you have an active web connection.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		info, err := getBinaryInfo()
		if err != nil {
			return err
		}
		release, err := getLatestRelease(info)
		if err != nil {
			return err
		}
		if info.version == release.version {
			fmt.Println("kube-cli is the latest version")
			return nil
		}
		// Download SHA512 sum file
		fmt.Println("Downloading SHA512 sum...")
		shaFile, err := downloadToTemp(release.shaURL)
		if err != nil {
			return err
		}
		defer os.Remove(shaFile)
		// Download tar.gz file
		fmt.Println("Downloading archive...")
		tarFile, err := downloadToTemp(release.tarURL)
		if err != nil {
			return err
		}
		defer os.Remove(tarFile)
		// Verify SHA512 sum
		downloadedSum, err := extractSum(shaFile)
		if err != nil {
			return err
		}
		computedSum, err := fileSha512Sum(tarFile)
		if err != nil {
			return err
		}
		if downloadedSum != computedSum {
			return errors.New("update failed, SHA512 sum missmatch")
		}
		fmt.Println("SHA512 verification succeeded")
		// Unarchive binary
		appName := filepath.Base(info.path)
		binaryFile, err := unarchiveBinary(tarFile, appName)
		if err != nil {
			return err
		}
		defer os.Remove(binaryFile)
		// Override existing binary
		err = os.Rename(binaryFile, info.path)
		if err != nil {
			return err
		}
		err = os.Chmod(info.path, 0755)
		if err != nil {
			return err
		}
		fmt.Println("Successfully updated tool from " + info.version + " to " + release.version)
		return nil
	},
}

func unarchiveBinary(path, name string) (string, error) {
	var tempName string
	f, err := ioutil.TempFile(os.TempDir(), "kube-cli-")
	if err != nil {
		return tempName, err
	}
	tempName = f.Name()
	ar, err := os.Open(path)
	if err != nil {
		return tempName, err
	}
	defer ar.Close()
	gzf, err := gzip.NewReader(ar)
	if err != nil {
		return tempName, err
	}
	tarReader := tar.NewReader(gzf)
	created := false
	for true {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return tempName, err
		}
		if header.Typeflag != tar.TypeReg {
			continue
		}
		if header.Name != name {
			continue
		}
		bin, err := os.Create(tempName)
		if err != nil {
			return tempName, err
		}
		defer bin.Close()
		io.Copy(bin, tarReader)
		created = true
	}
	if !created {
		return tempName, errors.New("unarchive failed, could not find the binary")
	}
	return tempName, nil
}

func extractSum(path string) (string, error) {
	var sum string
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return sum, err
	}
	sum = strings.Fields(string(b))[0]
	return sum, nil
}

func downloadToTemp(url string) (string, error) {
	var tempName string
	f, err := ioutil.TempFile(os.TempDir(), "kube-cli-")
	if err != nil {
		return tempName, err
	}
	tempName = f.Name()
	err = downloadFile(tempName, url)
	if err != nil {
		return tempName, err
	}
	return tempName, nil
}

func fileSha512Sum(path string) (string, error) {
	var shaSum string
	f, err := os.Open(path)
	if err != nil {
		return shaSum, err
	}
	defer f.Close()
	h := sha512.New()
	if _, err := io.Copy(h, f); err != nil {
		return shaSum, err
	}
	shaSum = fmt.Sprintf("%x", h.Sum(nil))
	return shaSum, nil
}

// WriteCounter counts the number of bytes written to it. It implements to the io.Writer
// interface and we can pass this into io.TeeReader() which will report progress on each
// write cycle.
type writeCounter struct {
	Total uint64
}

func (wc *writeCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Total += uint64(n)
	wc.PrintProgress()
	return n, nil
}

func (wc writeCounter) PrintProgress() {
	// Clear the line by using a character return to go back to the start and remove
	// the remaining characters by filling it with spaces
	fmt.Printf("\r%s", strings.Repeat(" ", 35))

	// Return again and print current status of download
	// We use the humanize package to print the bytes in a meaningful way (e.g. 10 MB)
	fmt.Printf("\rDownloading... %s complete", humanize.Bytes(wc.Total))
}

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory. We pass an io.TeeReader
// into Copy() to report progress on the download.
func downloadFile(filepath string, url string) error {

	// Create the file, but give it a tmp file extension, this means we won't overwrite a
	// file until it's downloaded, but we'll remove the tmp extension once downloaded.
	out, err := os.Create(filepath + ".tmp")
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create our progress reporter and pass it to be used alongside our writer
	counter := &writeCounter{}
	_, err = io.Copy(out, io.TeeReader(resp.Body, counter))
	if err != nil {
		return err
	}

	// The progress use the same line so print a new line once it's finished downloading
	fmt.Print("\n")

	err = os.Rename(filepath+".tmp", filepath)
	if err != nil {
		return err
	}

	return nil
}

type latestRelease struct {
	version string
	shaURL  string
	tarURL  string
}

func getLatestRelease(info binaryInfo) (latestRelease, error) {
	var release latestRelease
	var netTransport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
	}
	var netClient = &http.Client{
		Timeout:   time.Second * 10,
		Transport: netTransport,
	}
	var url = "https://api.github.com/repos/ajdnik/kube-cli/releases/latest"
	response, err := netClient.Get(url)
	if err != nil {
		return release, err
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return release, err
	}
	var raw map[string]interface{}
	json.Unmarshal(body, &raw)
	if _, ok := raw["tag_name"]; !ok {
		return release, errors.New("problem parsing json, missing tag_name property")
	}
	latestVersion := raw["tag_name"].(string)
	if _, ok := raw["assets"]; !ok {
		return release, errors.New("problem parsing json, missing assets property")
	}
	assets := raw["assets"].([]interface{})
	if len(assets) == 0 {
		return release, errors.New("problem parsing json, assets array is empty")
	}
	release = latestRelease{
		version: latestVersion,
	}
	appName := filepath.Base(info.path)
	shaName := appName + "_" + latestVersion + "_" + info.os + "_" + info.arch + ".sha512"
	tarName := appName + "_" + latestVersion + "_" + info.os + "_" + info.arch + ".tar.gz"
	for _, asset := range assets {
		a := asset.(map[string]interface{})
		if _, ok := a["name"]; !ok {
			continue
		}
		if _, ok := a["browser_download_url"]; !ok {
			continue
		}
		name := a["name"].(string)
		url := a["browser_download_url"].(string)
		if name == shaName {
			release.shaURL = url
		}
		if name == tarName {
			release.tarURL = url
		}
	}
	if len(release.shaURL) == 0 || len(release.tarURL) == 0 {
		return release, errors.New("problem parsing json, assets not found")
	}
	return release, nil
}

type binaryInfo struct {
	path    string
	arch    string
	os      string
	version string
}

func getBinaryInfo() (binaryInfo, error) {
	var info binaryInfo
	path, err := os.Executable()
	if err != nil {
		return info, err
	}
	info = binaryInfo{
		path:    path,
		arch:    runtime.GOARCH,
		os:      runtime.GOOS,
		version: version.Get(),
	}
	return info, nil
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
