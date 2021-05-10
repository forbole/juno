package types

import (
	"fmt"
	"io/ioutil"
	"path"

	"github.com/pelletier/go-toml"

	"github.com/rs/zerolog/log"
)

var (
	HomePath = ""
)

// GetConfigFilePath returns the path to the configuration file given the executable name
func GetConfigFilePath() string {
	return path.Join(HomePath, "config.toml")
}

// Read takes the path to a configuration file and returns the properly parsed configuration
func Read(configPath string, parser ConfigParser) (Config, error) {
	if configPath == "" {
		return nil, fmt.Errorf("empty configuration path")
	}

	log.Debug().Msg("reading config file")

	configData, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %s", err)
	}

	return parser(configData)
}

// Write allows to write the given configuration into the file present at the given path
func Write(cfg Config, path string) error {
	bz, err := toml.Marshal(cfg)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, bz, 0666)
}
