package parse

import (
	"fmt"
	"os"

	"github.com/forbole/juno/v3/types/config"

	"github.com/spf13/cobra"

	"github.com/forbole/juno/v3/types"
)

// ReadConfig represents a Cobra cmd function allowing to read the config before executing the command itself
func ReadConfig(cfg *Config) types.CobraCmdFunc {
	return func(_ *cobra.Command, _ []string) error {
		return UpdatedGlobalCfg(cfg)
	}
}

// UpdatedGlobalCfg parses the configuration file using the provided configuration and sets the
// parsed config as the global one
func UpdatedGlobalCfg(cfg *Config) error {
	file := config.GetConfigFilePath()

	// Make sure the path exists
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return fmt.Errorf("config file does not exist. Make sure you have run the init command")
	}

	// Read the config
	junoCfg, err := config.Read(file, cfg.GetConfigParser())
	if err != nil {
		return err
	}

	// Set the global configuration
	config.Cfg = junoCfg
	return nil
}
