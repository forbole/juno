package config

import (
	"io/ioutil"
	"path"

	"gopkg.in/yaml.v3"
)

var (
	HomePath = ""
)

// GetConfigFilePath returns the path to the configuration file given the executable name
func GetConfigFilePath() string {
	return path.Join(HomePath, "config.yaml")
}

// Write allows to write the given configuration into the file present at the given path
func Write(cfg Config, path string) error {
	bz, err := yaml.Marshal(&cfg)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, bz, 0666)
}
