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
