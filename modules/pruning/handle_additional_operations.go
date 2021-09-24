package pruning

import "fmt"

// RunAdditionalOperations runs the additional operations for the pruning module
func RunAdditionalOperations(cfg *Config) error {
	return checkConfig(cfg)
}

// checkConfig checks if the given config is valid
func checkConfig(cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("pruning config is not set but module is enabled")
	}

	return nil
}
