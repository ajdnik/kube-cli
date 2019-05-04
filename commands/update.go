package commands

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/ajdnik/kube-cli/executable"
	"github.com/ajdnik/kube-cli/hash"
	"github.com/ajdnik/kube-cli/tar"
	"github.com/ajdnik/kube-cli/web"
	humanize "github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
)

// UpdateCommand is a Cobra command that updates the executable from
// an online source.
var UpdateCommand = &cobra.Command{
	Use:   "update",
	Short: "Update the command line tool",
	Long: `Update the command line tool by pulling the latest
version from the web. Make sure you have an active web connection.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		info, err := executable.GetInfo()
		if err != nil {
			return err
		}
		shaName := info.Name + "_" + info.OS + "_" + info.Arch + ".sha512"
		tarName := info.Name + "_" + info.OS + "_" + info.Arch + ".tar.gz"
		release, err := web.GetLatestRelease(shaName, tarName)
		if err != nil {
			return err
		}
		if info.Version == release.Version {
			fmt.Println("kube-cli is the latest version")
			return nil
		}
		// Download SHA512 sum file
		fmt.Println("Downloading SHA512 sum...")
		shaTemp, err := createTempFile()
		if err != nil {
			return err
		}
		sz, err := web.DownloadFile(shaTemp, release.ShaURL)
		if err != nil {
			return err
		}
		defer os.Remove(shaTemp)
		fmt.Printf("Downloaded %s.", humanize.Bytes(uint64(sz)))
		fmt.Println("")
		// Download tar.gz file
		fmt.Println("Downloading archive...")
		tarTemp, err := createTempFile()
		if err != nil {
			return err
		}
		sz, err = web.DownloadFile(tarTemp, release.TarURL)
		if err != nil {
			return err
		}
		defer os.Remove(tarTemp)
		fmt.Printf("Downloaded %s.", humanize.Bytes(uint64(sz)))
		fmt.Println("")
		// Verify SHA512 sum
		downloadedSum, err := extractSum(shaTemp)
		if err != nil {
			return err
		}
		computedSum, err := hash.Sum(tarTemp)
		if err != nil {
			return err
		}
		if downloadedSum != computedSum {
			return errors.New("update failed, SHA512 sum missmatch")
		}
		fmt.Println("SHA512 verification succeeded")
		// Unarchive binary
		err = tar.Unarchive(tarTemp, filepath.Dir(info.Path))
		if err != nil {
			return err
		}
		fmt.Println("Successfully updated tool from " + info.Version + " to " + release.Version)
		return nil
	},
}

// Exists reports whether the named file exists.
func fileExists(name string) bool {
	info, err := os.Stat(name)
	if err != nil {
		return false
	}
	if info.IsDir() {
		return false
	}
	return true
}

func extractSum(path string) (string, error) {
	var sum string
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return sum, err
	}
	flds := strings.Fields(string(b))
	if len(flds) != 2 {
		return sum, errors.New("sum extraction failed, invalid format")
	}
	return flds[0], nil
}

func createTempFile() (string, error) {
	var name string
	f, err := ioutil.TempFile(os.TempDir(), "kube-cli-")
	if err != nil {
		return name, err
	}
	name = f.Name()
	return name, nil
}
