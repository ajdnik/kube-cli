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

// DeployCommand is a Cobra command that deploys the project to
// a Kubernetes cluster.
var DeployCommand = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy the project to Kubernetes",
	Long: `Deploy the project to Kubernetes cluster by building a
Docker image and deploying the image to a Kubernetes Deployment object.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		spin := ui.ShowSpinner(1, "Reading configuration...")
		cwd, err := executable.GetCwd()
		if err != nil {
			ui.SpinnerFail(1, "There was a problem reading configuration.", spin)
			ui.FailMessage("Please, retry 'kube-cli deploy' command.")
			return err
		}
		configPath, err := config.GetPath(cwd)
		if err != nil {
			ui.SpinnerFail(1, "There was a problem reading configuration.", spin)
			ui.FailMessage("Couldn't find kubecli YAML file in the project root. Try running 'kube-cli init' to create one.")
			return err
		}
		config, err := config.Read(configPath)
		if err != nil {
			ui.SpinnerFail(1, "There was a problem reading configuration.", spin)
			ui.FailMessage("Couldn't read kubecli YAML file. Try running 'kube-cli validate' to make sure the file is valid.")
			return err
		}
		ui.SpinnerSuccess(1, "Successfully read configuration for project.", spin)
		// Compress root of project into .tar.gz
		spin = ui.ShowSpinner(2, "Packing project into archive...")
		dockerfile := filepath.Join(cwd, "Dockerfile")
		if !filesystem.FileExists(dockerfile) {
			ui.SpinnerFail(2, "There was a problem packing a project into archive.", spin)
			ui.FailMessage("Couldn't find Dockerfile in the project root. Please add one.")
			return errors.New("missing Dockerfile")
		}
		tarTemp, err := filesystem.CreateTemp()
		if err != nil {
			ui.SpinnerFail(2, "There was a problem packing a project into archive.", spin)
			ui.FailMessage("Please, retry 'kube-cli deploy' command as an administrator.")
			return err
		}
		defer os.Remove(tarTemp)
		// Generate project files list
		projectFiles, err := filesystem.Glob(cwd)
		if err != nil {
			ui.SpinnerFail(2, "There was a problem packing a project into archive.", spin)
			ui.FailMessage("Please, retry 'kube-cli deploy' command as an administrator.")
			return err
		}
		filteredFiles, err := filterProjectFiles(projectFiles, cwd)
		if err != nil {
			ui.SpinnerFail(2, "There was a problem packing a project into archive.", spin)
			ui.FailMessage("Please, retry 'kube-cli deploy' command as an administrator.")
			return err
		}
		arTemp, err := filesystem.CreateTemp()
		if err != nil {
			ui.SpinnerFail(2, "There was a problem packing a project into archive.", spin)
			ui.FailMessage("Please, retry 'kube-cli deploy' command as an administrator.")
			return err
		}
		defer os.Remove(arTemp)
		err = tar.Archive(filteredFiles, arTemp, &cwd)
		if err != nil {
			ui.SpinnerFail(2, "There was a problem packing a project into archive.", spin)
			ui.FailMessage("Please, retry 'kube-cli deploy' command as an administrator.")
			return err
		}
		ui.SpinnerSuccess(2, "Project packing successfull.", spin)
		spin = ui.ShowSpinner(3, "Uploading archive...")
		bucketName := config.Gke.Project + "-cloudbuild"
		timestamp := fmt.Sprintf("%v", time.Now().Unix())
		objectName := fmt.Sprintf("%v-%v.tar.gz", config.Docker.Name, timestamp)
		_ = web.CreateBucket(bucketName, config.Gke.Project)
		sz, err := web.StorageUpload(bucketName, objectName, arTemp)
		if err != nil {
			ui.SpinnerFail(3, "There was a problem uploading archive.", spin)
			ui.FailMessage("Please, retry 'kube-cli deploy'. Make sure you have an active internet connection and 'Storage Admin' permissions on GCP Service Account defined in GOOGLE_APPLICATION_CREDENTIALS.")
			return err
		}
		ui.SpinnerSuccess(3, fmt.Sprintf("Uploaded archive %s.", humanize.Bytes(uint64(sz))), spin)
		spin = ui.ShowSpinner(4, "Building project...")
		tags := []string{
			"test",
			timestamp,
		}
		bld, err := web.CreateBuild(config.Gke.Project, config.Docker.Name, bucketName, objectName, tags)
		if err != nil {
			ui.SpinnerFail(4, "There was a problem building the project.", spin)
			ui.FailMessage("Please, retry 'kube-cli deploy'. Make sure you have an active internet connection and 'Cloud Build Service Account' permissions on GCP Service Account defined in GOOGLE_APPLICATION_CREDENTIALS.")
			return err
		}
		running := true
		timeout := 1
		for running {
			b, err := web.GetBuild(config.Gke.Project, bld.Id)
			if err != nil {
				ui.SpinnerFail(4, "There was a problem building the project.", spin)
				ui.FailMessage("Please, retry 'kube-cli deploy'. Make sure you have an active internet connection and 'Cloud Build Service Account' permissions on GCP Service Account defined in GOOGLE_APPLICATION_CREDENTIALS.")
				return err
			}
			if b.Status == web.SuccessBuildStatus {
				running = false
				break
			}
			if b.Status == web.QueuedBuildStatus || b.Status == web.WorkingBuildStatus {
				timeout *= 2
				time.Sleep(time.Duration(timeout) * time.Second)
				continue
			}
			ui.SpinnerFail(4, "There was a problem building the project.", spin)
			ui.FailMessage(fmt.Sprintf("There was a problem building the project, fix the issue and rerun the command. More info available at %v.", b.LogURL))
			return fmt.Errorf("visit %v to learn more", b.LogURL)
		}
		ui.SpinnerSuccess(4, "Building project succeeded.", spin)
		spin = ui.ShowSpinner(5, "Deploying project...")
		cls, err := web.GetGKECluster(config.Gke.Project, config.Gke.Zone, config.Gke.Cluster)
		if err != nil {
			ui.SpinnerFail(5, "There was a problem deploying the project.", spin)
			ui.FailMessage("Please, retry 'kube-cli deploy'. Make sure you have an active internet connection and 'Kubernetes Engine Admin' permissions on GCP Service Account defined in GOOGLE_APPLICATION_CREDENTIALS.")
			return err
		}
		dockerImage := fmt.Sprintf("gcr.io/%v/%v:%v", config.Gke.Project, config.Docker.Name, timestamp)
		err = web.UpdateDeployment(config.Deployment.Namespace, config.Deployment.Name, config.Deployment.Container.Name, dockerImage, cls)
		if err != nil {
			ui.SpinnerFail(5, "There was a problem deploying the project.", spin)
			ui.FailMessage("Please, retry 'kube-cli deploy'. Make sure you have an active internet connection and 'Kubernetes Engine Admin' permissions on GCP Service Account defined in GOOGLE_APPLICATION_CREDENTIALS.")
			return err
		}
		ui.SpinnerSuccess(5, "Deploying project succeeded.", spin)
		return nil
	},
}

func filterProjectFiles(files []string, cwd string) ([]string, error) {
	ignorePath := filepath.Join(cwd, ".kubecliignore")
	ignore := gitignore.New(strings.NewReader(".git/**/*\n.kubecliignore"), cwd, nil)
	var err error
	if filesystem.FileExists(ignorePath) {
		ignore, err = gitignore.NewFromFile(ignorePath)
		if err != nil {
			return files, err
		}
	}
	i := 0
	for _, f := range files {
		if ignore.Include(f) {
			files[i] = f
			i++
		}
	}
	return files[:i], nil
}
