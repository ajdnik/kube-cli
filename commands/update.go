package commands

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ajdnik/kube-cli/executable"
	"github.com/ajdnik/kube-cli/hash"
	"github.com/ajdnik/kube-cli/tar"
	"github.com/ajdnik/kube-cli/web"
	humanize "github.com/dustin/go-humanize"
	spinner "github.com/janeczku/go-spinner"
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
		spin := createNewSpinner(1, "Retrieving latest version info...")
		info, err := executable.GetInfo()
		if err != nil {
			spinnerFail(1, "There was a problem retrieving the latest version info.", spin)
			return err
		}
		shaName := info.Name + "_" + info.OS + "_" + info.Arch + ".sha512"
		tarName := info.Name + "_" + info.OS + "_" + info.Arch + ".tar.gz"
		release, err := web.GetLatestRelease(shaName, tarName)
		if err != nil {
			spinnerFail(1, "There was a problem retrieving the latest version info.", spin)
			return err
		}
		spinnerSuccess(1, fmt.Sprintf("Retrieved latest version is %v.", release.Version), spin)
		if info.Version == release.Version {
			fmt.Printf("The CLI tool is already updated to the latest version.")
			return nil
		}
		// Download tar.gz file
		spin = createNewSpinner(2, "Downloading CLI archive...")
		tarTemp, err := createTempFile()
		if err != nil {
			spinnerFail(2, "There was a problem downloading CLI archive.", spin)
			return err
		}
		sz, err := web.DownloadFile(tarTemp, release.TarURL)
		if err != nil {
			spinnerFail(2, "There was a problem downloading CLI archive.", spin)
			return err
		}
		defer os.Remove(tarTemp)
		spinnerSuccess(2, fmt.Sprintf("Downloaded CLI archive %s.", humanize.Bytes(uint64(sz))), spin)
		// Download SHA512 sum file
		spin = createNewSpinner(3, "Verifying downloaded archive...")
		shaTemp, err := createTempFile()
		if err != nil {
			spinnerFail(3, "There was a problem verifying the downloaded archive.", spin)
			return err
		}
		sz, err = web.DownloadFile(shaTemp, release.ShaURL)
		if err != nil {
			spinnerFail(3, "There was a problem verifying the downloaded archive.", spin)
			return err
		}
		defer os.Remove(shaTemp)
		// Verify SHA512 sum
		downloadedSum, err := extractSum(shaTemp)
		if err != nil {
			spinnerFail(3, "There was a problem verifying the downloaded archive.", spin)
			return err
		}
		computedSum, err := hash.Sum(tarTemp)
		if err != nil {
			spinnerFail(3, "There was a problem verifying the downloaded archive.", spin)
			return err
		}
		if downloadedSum != computedSum {
			spinnerFail(3, "There was a problem verifying the downloaded archive.", spin)
			return errors.New("update failed, SHA512 sum missmatch")
		}
		spinnerSuccess(3, "Verified downloaded archive.", spin)
		spin = createNewSpinner(4, "Updating CLI binaries...")
		// Unarchive binary
		err = tar.Unarchive(tarTemp, filepath.Dir(info.Path))
		if err != nil {
			spinnerFail(4, "There was a problem updating CLI binaries.", spin)
			return err
		}
		spinnerSuccess(4, fmt.Sprintf("Updated CLI binaries from %v to %v.", info.Version, release.Version), spin)
		return nil
	},
}

func createNewSpinner(step int8, descr string) *spinner.Spinner {
	spin := spinner.NewSpinner(fmt.Sprintf("Step %v: %v", step, descr))
	spin.SetSpeed(100 * time.Millisecond)
	spin.SetCharset([]string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"})
	spin.Start()
	return spin
}

func spinnerFail(step int8, descr string, spin *spinner.Spinner) {
	spin.Stop()
	fmt.Println(fmt.Sprintf("✖ Step %v: %v", step, descr))
}

func spinnerSuccess(step int8, descr string, spin *spinner.Spinner) {
	spin.Stop()
	fmt.Println(fmt.Sprintf("✓ Step %v: %v", step, descr))
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
