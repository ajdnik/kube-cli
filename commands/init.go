package commands

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"

	"github.com/ajdnik/kube-cli/config"
	"github.com/ajdnik/kube-cli/executable"
	"github.com/ajdnik/kube-cli/filesystem"
	"github.com/ajdnik/kube-cli/ui"
	"github.com/spf13/cobra"
)

// InitCommand generates a YAML config used by other
// commands to properly deploy the project to Kubernetes.
var InitCommand = &cobra.Command{
	Use:   "init",
	Short: "Initialize the project",
	Long: `Initialize the project YAML config by answering
some questions.`,
	RunE: func(cmd *cobra.Command, args []string) error {
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
		// Create new ignore file if user requests
		ip := filepath.Join(cwd, ".kubecliignore")
		if !filesystem.FileExists(ip) {
			create, err := ui.Confirm(".kubecliignore was not found in the project root. Would you like to create a generic one?")
			if err != nil {
				ui.FailMessage("Command canceled by user. No changes made.")
				return err
			}
			if create {
				err = ioutil.WriteFile(ip, []byte(".git/**/*\n.kubecliignore\nkubecli.yaml\nkubecli.yml"), 0644)
				if err != nil {
					ui.FailMessage("Please, retry 'kube-cli init' command. Try running it as an administrator.")
					return err
				}
				ui.SuccessMessage("Created a generic .kubecliignore file.")
			}
		}
		// Load existing YAML config, if it exists
		var cfg config.Data
		cp, err := config.GetPath(cwd)
		if err == nil {
			// Parse project YAML config
			cfg, err = config.Read(cp)
			if err != nil {
				ui.FailMessage("Couldn't read kubecli YAML file. Try running 'kube-cli lint' to make sure the file is valid.")
				return err
			}
			cont, err := ui.Confirm("Continuing will override the current YAML file. Are you sure?")
			if err != nil {
				ui.FailMessage("Command canceled by user. The configuration hasn't been modified.")
				return err
			}
			if !cont {
				ui.Message("The kubecli configuration hasn't been changed.")
				return nil
			}
		}
		// Prompt user for input to build YAML config
		ui.Message("Provide the following variables to build the project config file:")
		cfg.Gke.Project, err = ui.Ask("GKE Project", "Name of the GCP project where the Kubernetes cluster is hosted.", cfg.Gke.Project, validDashName)
		if err != nil {
			ui.FailMessage("Command canceled by user. No changes made.")
			return err
		}
		cfg.Gke.Cluster, err = ui.Ask("GKE Cluster", "Name of the GKE cluster.", cfg.Gke.Cluster, validDashName)
		if err != nil {
			ui.FailMessage("Command canceled by user. No changes made.")
			return err
		}
		cfg.Gke.Zone, err = ui.Choose("GKE Zone", "GCP zone of the Kubernetes cluster.", cfg.Gke.Zone, genZones())
		if err != nil {
			ui.FailMessage("Command canceled by user. No changes made.")
			return err
		}
		cfg.Docker.Name, err = ui.Ask("Docker Name", "Name of the Docker image, without gcr.io/...", cfg.Docker.Name, validDashName)
		if err != nil {
			ui.FailMessage("Command canceled by user. No changes made.")
			return err
		}
		cfg.Docker.Tag, err = ui.Ask("Docker Tag", "Name of a Docker tag that will be applied as a default.", cfg.Docker.Tag, validDashName)
		if err != nil {
			ui.FailMessage("Command canceled by user. No changes made.")
			return err
		}
		cfg.Deployment.Name, err = ui.Ask("Deployment Name", "Name of the Kubernetes deployment where the project is deployed.", cfg.Deployment.Name, validDashName)
		if err != nil {
			ui.FailMessage("Command canceled by user. No changes made.")
			return err
		}
		cfg.Deployment.Namespace, err = ui.Ask("Deployment Namespace", "Kubernetes namespace where the deplyment resides.", cfg.Deployment.Namespace, validDashName)
		if err != nil {
			ui.FailMessage("Command canceled by user. No changes made.")
			return err
		}
		cfg.Deployment.Container.Name, err = ui.Ask("Container Name", "Container name used in the Kubernetes deployment.", cfg.Deployment.Container.Name, validDashName)
		if err != nil {
			ui.FailMessage("Command canceled by user. No changes made.")
			return err
		}
		// Save config to YAML file
		err = config.Write(cp, cfg)
		if err != nil {
			ui.FailMessage("Couldn't save YAML config file. Please rerun the 'kube-cli init' command as an administrator.")
			return err
		}
		ui.SuccessMessage("Project is configured. You can now run 'kube-cli deploy' to deploy the project to Kubernetes.")
		return nil
	},
}

// Validate user input according to a regex rule.
func validDashName(input interface{}) error {
	// TODO: Improve name validation.
	r, _ := regexp.Compile("^[a-zA-Z][a-zA-Z0-9]*(-[a-zA-Z0-9]+)*$")
	if str, ok := input.(string); !ok || !r.MatchString(str) || len(str) < 3 {
		return errors.New("must be a string of alphanumeric characters delimited by dashes with a minimum length of 3 characters and it needs to start with an non numeric character, for example value-1")
	}
	return nil
}

// Generate GCP zone names from regions and supported zones.
func genZones() []string {
	regions := make(map[string][]string)
	regions["asia-east1"] = []string{"a", "b", "c"}
	regions["asia-east2"] = []string{"a", "b", "c"}
	regions["asia-northeast1"] = []string{"a", "b", "c"}
	regions["asia-northeast2"] = []string{"a", "b", "c"}
	regions["asia-south1"] = []string{"a", "b", "c"}
	regions["asia-southeast1"] = []string{"a", "b", "c"}
	regions["australia-southeast1"] = []string{"a", "b", "c"}
	regions["europe-north1"] = []string{"a", "b", "c"}
	regions["europe-west1"] = []string{"b", "c", "d"}
	regions["europe-west2"] = []string{"a", "b", "c"}
	regions["europe-west3"] = []string{"a", "b", "c"}
	regions["europe-west4"] = []string{"a", "b", "c"}
	regions["europe-west6"] = []string{"a", "b", "c"}
	regions["northamerica-northeast1"] = []string{"a", "b", "c"}
	regions["southamerica-east1"] = []string{"a", "b", "c"}
	regions["us-central1"] = []string{"a", "b", "c", "f"}
	regions["us-east1"] = []string{"b", "c", "d"}
	regions["us-east4"] = []string{"a", "b", "c"}
	regions["us-west1"] = []string{"a", "b", "c"}
	regions["us-west2"] = []string{"a", "b", "c"}
	var zones []string
	for reg, zz := range regions {
		for _, z := range zz {
			zones = append(zones, fmt.Sprintf("%v-%v", reg, z))
		}
	}
	return zones
}
