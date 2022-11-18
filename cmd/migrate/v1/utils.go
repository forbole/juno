package v1

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/pelletier/go-toml"

	"github.com/saifullah619/juno/v3/types/config"
)

// GetConfig returns the configuration reading it from the config.toml file present inside the home directory
func GetConfig() (Config, error) {
	file := path.Join(config.HomePath, "config.toml")

	// Make sure the path exists
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return Config{}, fmt.Errorf("config file does not exist")
	}

	bz, err := ioutil.ReadFile(file)
	if err != nil {
		return Config{}, fmt.Errorf("error while reading config files: %s", err)
	}

	var cfg Config
	err = toml.Unmarshal(bz, &cfg)
	return cfg, err
}
