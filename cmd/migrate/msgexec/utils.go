package msgexec

import (
	"fmt"
	"os"
	"path"

	utils "github.com/forbole/juno/v5/cmd/migrate/utils"
	"gopkg.in/yaml.v3"

	"github.com/forbole/juno/v5/types/config"
)

// GetConfig returns the configuration reading it from the config.yaml file present inside the home directory
func GetConfig() (utils.Config, error) {
	file := path.Join(config.HomePath, "config.yaml")

	// Make sure the path exists
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return utils.Config{}, fmt.Errorf("config file does not exist")
	}

	bz, err := os.ReadFile(file)
	if err != nil {
		return utils.Config{}, fmt.Errorf("error while reading config files: %s", err)
	}

	var cfg utils.Config
	err = yaml.Unmarshal(bz, &cfg)
	return cfg, err
}