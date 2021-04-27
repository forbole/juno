package parse

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/desmos-labs/juno/types"
)

// ReadConfig parses the configuration file for the executable having the give name using
// the provided configuration parser
func ReadConfig(cfg *Config) types.CobraCmdFunc {
	return func(_ *cobra.Command, _ []string) error {
		file := types.GetConfigFilePath(cfg.GetName())

		// Make sure the path exists
		if _, err := os.Stat(file); os.IsNotExist(err) {
			return fmt.Errorf("%s file does not exist. Make sure you have run %s init", file, cfg.GetName())
		}

		// Read the config
		cfg, err := types.Read(file, cfg.GetConfigParser())
		if err != nil {
			return err
		}

		// Set the global configuration
		types.Cfg = cfg
		return nil
	}
}
