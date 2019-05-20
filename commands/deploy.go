package commands

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ajdnik/kube-cli/config"
	"github.com/ajdnik/kube-cli/executable"
	"github.com/ajdnik/kube-cli/filesystem"
	"github.com/ajdnik/kube-cli/tar"
	"github.com/ajdnik/kube-cli/ui"
	"github.com/ajdnik/kube-cli/web"
	gitignore "github.com/denormal/go-gitignore"
	humanize "github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
)

// DeployCommand executes a multi step workflow that builds the
// project using GCP Cloud Build and than deploys the docker image
// to a GKE deployment.
var DeployCommand = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy the project to Kubernetes",
	Long: `Deploy the project to Kubernetes cluster by building a
Docker image and deploying the image to a Kubernetes Deployment object.`,
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
		spin = ui.ShowSpinner(2, "Packing project into archive...")
		// Verify Dockerfile exists in project root
		df := filepath.Join(cwd, "Dockerfile")
		if !filesystem.FileExists(df) {
			ui.SpinnerFail(2, "There was a problem packing a project into archive.", spin)
			ui.FailMessage("Couldn't find Dockerfile in the project root. Please add one.")
			return errors.New("missing Dockerfile")
		}
		// Generate project files list
		pf, err := filesystem.Glob(cwd)
		if err != nil {
			ui.SpinnerFail(2, "There was a problem packing a project into archive.", spin)
			ui.FailMessage("Please, retry 'kube-cli deploy' command as an administrator.")
			return err
		}
		// Remove files filtered by .kubecliignore file
		files, err := filterProjectFiles(pf, cwd)
		if err != nil {
			ui.SpinnerFail(2, "There was a problem packing a project into archive.", spin)
			ui.FailMessage("Please, retry 'kube-cli deploy' command as an administrator.")
			return err
		}
		// Create temp file for project archive
		tmp, err := filesystem.CreateTemp()
		if err != nil {
			ui.SpinnerFail(2, "There was a problem packing a project into archive.", spin)
			ui.FailMessage("Please, retry 'kube-cli deploy' command as an administrator.")
			return err
		}
		defer os.Remove(tmp)
		// Create .tar.gz archive from project files
		err = tar.Archive(files, tmp, &cwd)
		if err != nil {
			ui.SpinnerFail(2, "There was a problem packing a project into archive.", spin)
			ui.FailMessage("Please, retry 'kube-cli deploy' command as an administrator.")
			return err
		}
		ui.SpinnerSuccess(2, "Project packing successfull.", spin)
		spin = ui.ShowSpinner(3, "Uploading archive...")
		// Upload .tar.gz archive to GCP Storage
		bName := cfg.Gke.Project + "-cloudbuild"
		timestamp := fmt.Sprintf("%v", time.Now().Unix())
		oName := fmt.Sprintf("%v-%v.tar.gz", cfg.Docker.Name, timestamp)
		_ = web.CreateBucket(bName, cfg.Gke.Project)
		sz, err := web.StorageUpload(bName, oName, tmp)
		if err != nil {
			ui.SpinnerFail(3, "There was a problem uploading archive.", spin)
			ui.FailMessage("Please, retry 'kube-cli deploy'. Make sure you have an active internet connection and 'Storage Admin' permissions on GCP Service Account defined in GOOGLE_APPLICATION_CREDENTIALS.")
			return err
		}
		ui.SpinnerSuccess(3, fmt.Sprintf("Uploaded archive %s.", humanize.Bytes(uint64(sz))), spin)
		spin = ui.ShowSpinner(4, "Building project...")
		// Create build on GCP Cloud Build
		tags := []string{
			"test",
			timestamp,
		}
		bld, err := web.CreateBuild(cfg.Gke.Project, cfg.Docker.Name, bName, oName, tags)
		if err != nil {
			ui.SpinnerFail(4, "There was a problem building the project.", spin)
			ui.FailMessage("Please, retry 'kube-cli deploy'. Make sure you have an active internet connection and 'Cloud Build Service Account' permissions on GCP Service Account defined in GOOGLE_APPLICATION_CREDENTIALS.")
			return err
		}
		// Periodically check on build status
		running := true
		timeout := 1
		for running {
			b, err := web.GetBuild(cfg.Gke.Project, bld.ID)
			if err != nil {
				ui.SpinnerFail(4, "There was a problem building the project.", spin)
				ui.FailMessage("Please, retry 'kube-cli deploy'. Make sure you have an active internet connection and 'Cloud Build Service Account' permissions on GCP Service Account defined in GOOGLE_APPLICATION_CREDENTIALS.")
				return err
			}
			// The build succeeded
			if b.Status == web.SuccessBuildStatus {
				running = false
				break
			}
			// The build is still running or waiting to be run
			if b.Status == web.QueuedBuildStatus || b.Status == web.WorkingBuildStatus {
				timeout *= 2
				time.Sleep(time.Duration(timeout) * time.Second)
				continue
			}
			// The build failed
			ui.SpinnerFail(4, "There was a problem building the project.", spin)
			ui.FailMessage(fmt.Sprintf("There was a problem building the project, fix the issue and rerun the command. More info available at %v.", b.LogURL))
			return fmt.Errorf("visit %v to learn more", b.LogURL)
		}
		ui.SpinnerSuccess(4, "Building project succeeded.", spin)
		spin = ui.ShowSpinner(5, "Deploying project...")
		// Retrieve GKE cluster info
		cls, err := web.GetGKECluster(cfg.Gke.Project, cfg.Gke.Zone, cfg.Gke.Cluster)
		if err != nil {
			ui.SpinnerFail(5, "There was a problem deploying the project.", spin)
			ui.FailMessage("Please, retry 'kube-cli deploy'. Make sure you have an active internet connection and 'Kubernetes Engine Admin' permissions on GCP Service Account defined in GOOGLE_APPLICATION_CREDENTIALS.")
			return err
		}
		// Update GKE deployment image
		di := fmt.Sprintf("gcr.io/%v/%v:%v", cfg.Gke.Project, cfg.Docker.Name, timestamp)
		err = web.UpdateDeployment(cfg.Deployment.Namespace, cfg.Deployment.Name, cfg.Deployment.Container.Name, di, cls)
		if err != nil {
			ui.SpinnerFail(5, "There was a problem deploying the project.", spin)
			ui.FailMessage("Please, retry 'kube-cli deploy'. Make sure you have an active internet connection and 'Kubernetes Engine Admin' permissions on GCP Service Account defined in GOOGLE_APPLICATION_CREDENTIALS.")
			return err
		}
		ui.SpinnerSuccess(5, "Deploying project succeeded.", spin)
		return nil
	},
}

// Filter files based on rules defined in .kubecliignore file.
func filterProjectFiles(files []string, cwd string) ([]string, error) {
	ip := filepath.Join(cwd, ".kubecliignore")
	ignore := gitignore.New(strings.NewReader(".git/**/*\n.kubecliignore"), cwd, nil)
	var err error
	// Load .kubecliignore if it exists
	if filesystem.FileExists(ip) {
		ignore, err = gitignore.NewFromFile(ip)
		if err != nil {
			return files, err
		}
	}
	// Keep files that aren(t filtered by ignore rules
	i := 0
	for _, f := range files {
		if ignore.Include(f) {
			files[i] = f
			i++
		}
	}
	return files[:i], nil
}
