package config

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ajdnik/kube-cli/filesystem"
	yaml "gopkg.in/yaml.v2"
)

// Data represents the configuration structure of kubecli.yaml file.
type Data struct {
	Gke        GKEData
	Docker     DockerData
	Deployment DeploymentData
}

// GKEData represents the gke subsection of the kubecli.yaml file.
type GKEData struct {
	Project string
	Zone    string
	Cluster string
}

// DockerData represents the docker subsection of the kubecli.yaml file.
type DockerData struct {
	Name string
}

// DeploymentData represents the deployment subsection of the kubecli.yaml file.
type DeploymentData struct {
	Name      string
	Namespace string
	Container ContainerData
}

// ContainerData represents the container subsection of the kubecli.yaml file.
type ContainerData struct {
	Name string
}

// Read kubecli.yaml file and parse it.
func Read(file string) (Data, error) {
	var data Data
	f, err := os.Open(file)
	if err != nil {
		return data, err
	}
	defer f.Close()
	bytes, err := ioutil.ReadAll(f)
	if err != nil {
		return data, err
	}
	err = yaml.Unmarshal(bytes, &data)
	if err != nil {
		return data, err
	}
	return data, nil
}

// GetPath tries to return a valid .kubecli.yaml file path.
func GetPath(cwd string) (string, error) {
	configPathYaml := filepath.Join(cwd, "kubecli.yaml")
	if filesystem.FileExists(configPathYaml) {
		return configPathYaml, nil
	}
	configPathYml := filepath.Join(cwd, "kubecli.yml")
	if filesystem.FileExists(configPathYml) {
		return configPathYml, nil
	}
	return configPathYml, errors.New("kubecli yaml file not found")
}
