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
	"github.com/ajdnik/kube-cli/ui"
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
		spin := ui.ShowSpinner(1, "Retrieving latest version info...")
		info, err := executable.GetInfo()
		if err != nil {
			ui.SpinnerFail(1, "There was a problem retrieving the latest version info.", spin)
			return err
		}
		shaName := info.Name + "_" + info.OS + "_" + info.Arch + ".sha512"
		tarName := info.Name + "_" + info.OS + "_" + info.Arch + ".tar.gz"
		release, err := web.GetLatestRelease(shaName, tarName)
		if err != nil {
			ui.SpinnerFail(1, "There was a problem retrieving the latest version info.", spin)
			return err
		}
		ui.SpinnerSuccess(1, fmt.Sprintf("Retrieved latest version is %v.", release.Version), spin)
		if info.Version == release.Version {
			ui.SuccessMessage("The CLI tool is already updated to the latest version.")
			return nil
		}
		// Download tar.gz file
		spin = ui.ShowSpinner(2, "Downloading CLI archive...")
		tarTemp, err := createTempFile()
		if err != nil {
			ui.SpinnerFail(2, "There was a problem downloading CLI archive.", spin)
			return err
		}
		sz, err := web.DownloadFile(tarTemp, release.TarURL)
		if err != nil {
			ui.SpinnerFail(2, "There was a problem downloading CLI archive.", spin)
			return err
		}
		defer os.Remove(tarTemp)
		ui.SpinnerSuccess(2, fmt.Sprintf("Downloaded CLI archive %s.", humanize.Bytes(uint64(sz))), spin)
		// Download SHA512 sum file
		spin = ui.ShowSpinner(3, "Verifying downloaded archive...")
		shaTemp, err := createTempFile()
		if err != nil {
			ui.SpinnerFail(3, "There was a problem verifying the downloaded archive.", spin)
			return err
		}
		sz, err = web.DownloadFile(shaTemp, release.ShaURL)
		if err != nil {
			ui.SpinnerFail(3, "There was a problem verifying the downloaded archive.", spin)
			return err
		}
		defer os.Remove(shaTemp)
		// Verify SHA512 sum
		downloadedSum, err := extractSum(shaTemp)
		if err != nil {
			ui.SpinnerFail(3, "There was a problem verifying the downloaded archive.", spin)
			return err
		}
		computedSum, err := hash.Sum(tarTemp)
		if err != nil {
			ui.SpinnerFail(3, "There was a problem verifying the downloaded archive.", spin)
			return err
		}
		if downloadedSum != computedSum {
			ui.SpinnerFail(3, "There was a problem verifying the downloaded archive.", spin)
			return errors.New("update failed, SHA512 sum missmatch")
		}
		ui.SpinnerSuccess(3, "Verified downloaded archive.", spin)
		spin = ui.ShowSpinner(4, "Updating CLI binaries...")
		// Unarchive binary
		err = tar.Unarchive(tarTemp, filepath.Dir(info.Path))
		if err != nil {
			ui.SpinnerFail(4, "There was a problem updating CLI binaries.", spin)
			return err
		}
		ui.SpinnerSuccess(4, fmt.Sprintf("Updated CLI binaries from %v to %v.", info.Version, release.Version), spin)
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
