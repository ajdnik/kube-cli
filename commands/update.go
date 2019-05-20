package commands

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/ajdnik/kube-cli/executable"
	"github.com/ajdnik/kube-cli/filesystem"
	"github.com/ajdnik/kube-cli/hash"
	"github.com/ajdnik/kube-cli/tar"
	"github.com/ajdnik/kube-cli/ui"
	"github.com/ajdnik/kube-cli/web"
	humanize "github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
)

// UpdateCommand executes CLI update workflow, which downloads
// latest CLI tool, verifies the download and replaces the old
// binary.
var UpdateCommand = &cobra.Command{
	Use:   "update",
	Short: "Update the command line tool",
	Long: `Update the command line tool by pulling the latest
version from the web. Make sure you have an active web connection.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		spin := ui.ShowSpinner(1, "Retrieving latest version info...")
		// Get CLI binary info
		info, err := executable.GetInfo()
		if err != nil {
			ui.SpinnerFail(1, "There was a problem retrieving the latest version info.", spin)
			ui.FailMessage("Please, retry 'kube-cli update' command.")
			return err
		}
		// Retrieve latest version info
		shaName := info.Name + "_" + info.OS + "_" + info.Arch + ".sha512"
		tarName := info.Name + "_" + info.OS + "_" + info.Arch + ".tar.gz"
		release, err := web.GetLatestRelease(shaName, tarName)
		if err != nil {
			ui.SpinnerFail(1, "There was a problem retrieving the latest version info.", spin)
			ui.FailMessage("Please, retry 'kube-cli update' command. Make sure you have an active internet connection.")
			return err
		}
		ui.SpinnerSuccess(1, fmt.Sprintf("Retrieved latest version is %v.", release.Version), spin)
		if info.Version == release.Version {
			ui.SuccessMessage("The CLI tool is already updated to the latest version.")
			return nil
		}
		spin = ui.ShowSpinner(2, "Downloading CLI archive...")
		// Create temp file for downloaded archive
		tarTemp, err := filesystem.CreateTemp()
		if err != nil {
			ui.SpinnerFail(2, "There was a problem downloading CLI archive.", spin)
			ui.FailMessage("Please, retry 'kube-cli update' command as an administrator.")
			return err
		}
		// Download latest CLI archive
		sz, err := web.DownloadFile(tarTemp, release.TarURL)
		if err != nil {
			ui.SpinnerFail(2, "There was a problem downloading CLI archive.", spin)
			ui.FailMessage("Please, retry 'kube-cli update' command as an administrator.")
			return err
		}
		defer os.Remove(tarTemp)
		ui.SpinnerSuccess(2, fmt.Sprintf("Downloaded CLI archive %s.", humanize.Bytes(uint64(sz))), spin)
		// Download SHA512 sum file
		spin = ui.ShowSpinner(3, "Verifying downloaded archive...")
		// Create temp file for downloaded hash file
		shaTemp, err := filesystem.CreateTemp()
		if err != nil {
			ui.SpinnerFail(3, "There was a problem verifying the downloaded archive.", spin)
			ui.FailMessage("Please, retry 'kube-cli update' command as an administrator.")
			return err
		}
		// Download archive hash
		sz, err = web.DownloadFile(shaTemp, release.ShaURL)
		if err != nil {
			ui.SpinnerFail(3, "There was a problem verifying the downloaded archive.", spin)
			ui.FailMessage("Please, retry 'kube-cli update' command as an administrator.")
			return err
		}
		defer os.Remove(shaTemp)
		// Extract downloaded hash from file
		dSum, err := extractSum(shaTemp)
		if err != nil {
			ui.SpinnerFail(3, "There was a problem verifying the downloaded archive.", spin)
			ui.FailMessage("Please, retry 'kube-cli update' command.")
			return err
		}
		// Compute hash from downloaded archive
		cSum, err := hash.Sum(tarTemp)
		if err != nil {
			ui.SpinnerFail(3, "There was a problem verifying the downloaded archive.", spin)
			ui.FailMessage("Please, retry 'kube-cli update' command.")
			return err
		}
		// Verify hash sums match
		if dSum != cSum {
			ui.SpinnerFail(3, "There was a problem verifying the downloaded archive.", spin)
			ui.FailMessage("Please, retry 'kube-cli update' command. The downloaded archive was corrupt.")
			return errors.New("update failed, SHA512 sum missmatch")
		}
		ui.SpinnerSuccess(3, "Verified downloaded archive.", spin)
		spin = ui.ShowSpinner(4, "Updating CLI binaries...")
		// Unarchive binary
		err = tar.Unarchive(tarTemp, filepath.Dir(info.Path))
		if err != nil {
			ui.SpinnerFail(4, "There was a problem updating CLI binaries.", spin)
			ui.FailMessage("Please, retry 'kube-cli update' command as an administrator.")
			return err
		}
		ui.SpinnerSuccess(4, fmt.Sprintf("Updated CLI binaries from %v to %v.", info.Version, release.Version), spin)
		return nil
	},
}

// Extract hash sum from a .sha512 file.
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
