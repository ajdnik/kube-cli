package commands

import (
	"time"

	"github.com/ajdnik/kube-cli/config"
	"github.com/ajdnik/kube-cli/executable"
	"github.com/ajdnik/kube-cli/ui"
	"github.com/ajdnik/kube-cli/web"
	"github.com/spf13/cobra"
)

var asyncRollback bool

// RollbackCommand rolls back a Kubernetes deployment to a previous state.
var RollbackCommand = &cobra.Command{
	Use:   "rollback",
	Short: "Rollback deployment",
	Long:  `Rollback deployment to a previous state.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		spin := ui.ShowSpinner(1, "Reading configuration...")
		// Get project root directory
		cwd, err := executable.GetCwd()
		if err != nil {
			ui.SpinnerFail(1, "There was a problem reading configuration.", spin)
			ui.FailMessage("Please, retry 'kube-cli deploy' command.")
			return err
		}
		// Get YAML config path in project root
		cp, err := config.GetPath(cwd)
		if err != nil {
			ui.SpinnerFail(1, "There was a problem reading configuration.", spin)
			ui.FailMessage("Couldn't find kubecli YAML file in the project root. Try running 'kube-cli init' to create one.")
			return err
		}
		// Parse project YAML config
		cfg, err := config.Read(cp)
		if err != nil {
			ui.SpinnerFail(1, "There was a problem reading configuration.", spin)
			ui.FailMessage("Couldn't read kubecli YAML file. Try running 'kube-cli validate' to make sure the file is valid.")
			return err
		}
		ui.SpinnerSuccess(1, "Successfully read configuration for project.", spin)
		spin = ui.ShowSpinner(2, "Rolling back deployment...")
		// Retrieve GKE cluster info
		cls, err := web.GetGKECluster(cfg.Gke.Project, cfg.Gke.Zone, cfg.Gke.Cluster)
		if err != nil {
			ui.SpinnerFail(2, "There was a problem rolling back the deployment.", spin)
			ui.FailMessage("Please, retry 'kube-cli rollback'. Make sure you have an active internet connection and 'Kubernetes Engine Admin' permissions on GCP Service Account defined in GOOGLE_APPLICATION_CREDENTIALS.")
			return err
		}
		// Rollback deployment
		err = web.RollbackDeployment(cfg.Deployment.Namespace, cfg.Deployment.Name, cls)
		if err != nil {
			ui.SpinnerFail(2, "There was a problem rolling back the deployment.", spin)
			ui.FailMessage("Please, retry 'kube-cli rollback'. Make sure you have an active internet connection and 'Kubernetes Engine Admin' permissions on GCP Service Account defined in GOOGLE_APPLICATION_CREDENTIALS.")
			return err
		}
		if asyncRollback {
			ui.SpinnerSuccess(2, "Successfully started the rollback of the deployment. You can keep track of the progress at https://console.cloud.google.com/kubernetes/workload.", spin)
			return nil
		}
		// Periodically check deployment
		running := true
		timeout := 1
		maxTimeout := 60
		for running {
			cnt, err := web.UnavailableReplicas(cfg.Deployment.Namespace, cfg.Deployment.Name, cls)
			if err != nil {
				ui.SpinnerFail(2, "There was a problem rolling back the deployment.", spin)
				ui.FailMessage("Something unexpected happened. Please check on the status of the rollback on the Google Cloud Console https://console.cloud.google.com/kubernetes/workload.")
				return err
			}
			if cnt == 0 {
				running = false
				break
			}
			timeout *= 2
			if timeout > maxTimeout {
				timeout = maxTimeout
			}
			time.Sleep(time.Duration(timeout) * time.Second)
		}
		ui.SpinnerSuccess(2, "Successfully rolled back deployment.", spin)
		return nil
	},
}

func init() {
	RollbackCommand.Flags().BoolVarP(&asyncRollback, "async", "a", false, "Don't wait for rollback operation to complete")
}
