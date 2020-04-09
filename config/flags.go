package config

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/cli"
)

const (
	FlagStartHeight     = "start-height"
	FlagWorkerCount     = "workers"
	FlagListenNewBlocks = "listen-new-blocks"
	FlagLogLevel        = "log-level"
	FlagLogFormat       = "log-format"
	FlagFormat          = "format"
)

// PrepareMainCmd is meant to prepare the given command binding all the viper flags
func PrepareMainCmd(cmd *cobra.Command) cli.Executor {
	cmd.PersistentPreRunE = concatCobraCmdFuncs(bindFlagsLoadViper, cmd.PersistentPreRunE)
	return cli.Executor{Command: cmd, Exit: os.Exit}
}

// Bind all flags and read the config into viper
func bindFlagsLoadViper(cmd *cobra.Command, _ []string) error {
	// cmd.Flags() includes flags from this command and all persistent flags from the parent
	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		return err
	}

	return nil
}

type cobraCmdFunc func(cmd *cobra.Command, args []string) error

// Returns a single function that calls each argument function in sequence
// RunE, PreRunE, PersistentPreRunE, etc. all have this same signature
func concatCobraCmdFuncs(fs ...cobraCmdFunc) cobraCmdFunc {
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
