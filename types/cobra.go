package types

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// CobraCmdFunc represents a cobra command function
type CobraCmdFunc func(cmd *cobra.Command, args []string) error

// ConcatCobraCmdFuncs returns a single function that calls each argument function in sequence
// RunE, PreRunE, PersistentPreRunE, etc. all have this same signature
func ConcatCobraCmdFuncs(fs ...CobraCmdFunc) CobraCmdFunc {
	return func(cmd *cobra.Command, args []string) error {
		for _, f := range fs {
			if f != nil {
				if err := f(cmd, args); err != nil {
					return err
				}
			}
		}
		return nil
	}
}

// BindFlagsLoadViper binds all flags and read the config into viper
func BindFlagsLoadViper(cmd *cobra.Command, _ []string) error {
	// cmd.Flags() includes flags from this command and all persistent flags from the parent
	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		return err
	}

	return nil
}
