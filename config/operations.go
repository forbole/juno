package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/pelletier/go-toml"

	"github.com/rs/zerolog/log"
)

var (
	HomeDir, _ = os.UserHomeDir()
)

func GetConfigFolderPath(name string) string {
	return path.Join(HomeDir, fmt.Sprintf(".%s", name))
}

func GetConfigFilePath(name string) string {
	return path.Join(GetConfigFolderPath(name), "config.toml")
}

// SetupConfig takes the path to a configuration file and returns the properly parsed configuration
func Read(configPath string) (*Config, error) {
	if configPath == "" {
		return nil, fmt.Errorf("empty configuration path")
	}

	log.Debug().Msg("reading config file")

	configData, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %s", err)
	}

	return ParseString(configData)
}

// ParseString attempts to read and parse a Juno config from the given string bytes.
// An error reading or parsing the config results in a panic.
func ParseString(configData []byte) (*Config, error) {
	var cfg Config

	err := toml.Unmarshal(configData, &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to decode config: %s", err)
	}

	return &cfg, nil
}

// Write allows to write the given configuration into the file present at the given path
func Write(cfg *Config, path string) error {
	bz, err := toml.Marshal(cfg)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, bz, 0666)
}
