package parse

import (
	"fmt"
	"os"

	"github.com/forbole/juno/v2/types/config"

	"github.com/spf13/cobra"

	"github.com/forbole/juno/v2/types"
)

// ReadConfig parses the configuration file for the executable having the give name using
// the provided configuration parser
func ReadConfig(cfg *Config) types.CobraCmdFunc {
	return func(_ *cobra.Command, _ []string) error {
		file := config.GetConfigFilePath()

		// Make sure the path exists
		if _, err := os.Stat(file); os.IsNotExist(err) {
			return fmt.Errorf("config file does not exist. Make sure you have run the init command")
		}

		// Read the config
		cfg, err := config.Read(file, cfg.GetConfigParser())
		if err != nil {
			return err
		}

		// Set the global configuration
		config.Cfg = cfg
		return nil
	}
}
