package config

import (
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

// Write YAML config to file.
func Write(file string, data Data) error {
	str, err := yaml.Marshal(data)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(file, str, 0644)
}
