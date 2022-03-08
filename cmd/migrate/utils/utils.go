package utils

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/forbole/juno/v2/types/config"
)

// GetConfigFilePath returns the path to the configuration file given the executable name
func GetConfigFilePath() string {
	return path.Join(config.HomePath, "config.toml")
}

// ReadConfig reads the config.toml file contents
func ReadConfig() ([]byte, error) {
	file := GetConfigFilePath()

	// Make sure the path exists
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file does not exist")
	}

	bz, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("error while reading config files: %s", err)
	}

	return bz, nil
}
