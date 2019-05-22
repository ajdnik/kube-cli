package commands

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ajdnik/kube-cli/config"
	"github.com/ajdnik/kube-cli/executable"
	"github.com/ajdnik/kube-cli/filesystem"
	"github.com/ajdnik/kube-cli/ui"
	"github.com/spf13/cobra"
)

// ValidateCommand checks the validity of kubecli.yaml project config.
var ValidateCommand = &cobra.Command{
	Use:   "validate",
	Short: "Validate YAML config",
	Long: `Validate YAML config in kubecli.yaml to make sure
the properties and structure is correct.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Display warning if Dockerfile doesn't exist
		// Check if kubecli.yaml exists
		// Validate config
		// Validate ignore file if it exists
		// Get project root directory
		cwd, err := executable.GetCwd()
		if err != nil {
			ui.FailMessage("Please, retry 'kube-cli init' command.")
			return err
		}
		// Verify Dockerfile exists in project root
		df := filepath.Join(cwd, "Dockerfile")
		if !filesystem.FileExists(df) {
			ui.WarnMessage("Couldn't find Dockerfile in the project root. See https://docs.docker.com/engine/reference/builder/ for further info.")
		}
		// Get YAML config path in project root
		cp, err := config.GetPath(cwd)
		if err != nil {
			ui.FailMessage("Couldn't find kubecli YAML file in the project root. Try running 'kube-cli init' to create one.")
			return err
		}
		// Parse project YAML config
		cfg, err := config.Read(cp)
		if err != nil {
			ui.FailMessage(strings.Replace(err.Error(), "yaml:", "YAML sytnax is incorrect on", 1))
			return err
		}
		// Validate each property
		err = validDashName(cfg.Gke.Project)
		hasInvalid := false
		if err != nil {
			ui.FailMessage(fmt.Sprintf("GKE Project %v", err.Error()))
			hasInvalid = true
		}
		err = validDashName(cfg.Gke.Cluster)
		if err != nil {
			ui.FailMessage(fmt.Sprintf("GKE Cluster %v", err.Error()))
			hasInvalid = true
		}
		if !linearSearch(cfg.Gke.Zone, genZones()) {
			ui.FailMessage("GKE Zone is not a valid zone string. See https://cloud.google.com/compute/docs/regions-zones/ for more info.")
			hasInvalid = true
		}
		err = validDashName(cfg.Docker.Name)
		if err != nil {
			ui.FailMessage(fmt.Sprintf("Docker Name %v", err.Error()))
			hasInvalid = true
		}
		err = validDashName(cfg.Docker.Tag)
		if err != nil {
			ui.FailMessage(fmt.Sprintf("Docker Tag %v", err.Error()))
			hasInvalid = true
		}
		err = validDashName(cfg.Deployment.Name)
		if err != nil {
			ui.FailMessage(fmt.Sprintf("Deployment Name %v", err.Error()))
			hasInvalid = true
		}
		err = validDashName(cfg.Deployment.Namespace)
		if err != nil {
			ui.FailMessage(fmt.Sprintf("Deployment Namespace %v", err.Error()))
			hasInvalid = true
		}
		err = validDashName(cfg.Deployment.Container.Name)
		if err != nil {
			ui.FailMessage(fmt.Sprintf("Container Name %v", err.Error()))
			hasInvalid = true
		}
		if hasInvalid {
			ui.FailMessage("YAML configuration is invalid. Try running 'kube-cli init' to fix it.")
			return err
		}
		ui.SuccessMessage("YAML configuration is valid.")
		return nil
	},
}

// Linear search through a slice of strings.
func linearSearch(item string, arr []string) bool {
	for _, s := range arr {
		if item == s {
			return true
		}
	}
	return false
}
