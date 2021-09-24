package config

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

// Parser represents a function that allows to parse a file contents as a Config object
type Parser = func(fileContents []byte) (Config, error)

// DefaultConfigParser attempts to read and parse a Juno config from the given string bytes.
// An error reading or parsing the config results in a panic.
func DefaultConfigParser(configData []byte) (Config, error) {
	var cfg = Config{
		bytes: configData,
	}
	err := yaml.Unmarshal(configData, &cfg)
	return cfg, err
}

// Read takes the path to a configuration file and returns the properly parsed configuration
func Read(configPath string, parser Parser) (Config, error) {
	if configPath == "" {
		return Config{}, fmt.Errorf("empty configuration path")
	}

	configData, err := ioutil.ReadFile(configPath)
	if err != nil {
		return Config{}, fmt.Errorf("failed to read config: %s", err)
	}

	return parser(configData)
}
