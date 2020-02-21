package main

import (
	"os"

	"github.com/angelorc/desmos-parser/parse"
	"github.com/angelorc/desmos-parser/version"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/desmos-labs/desmos/app"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/cli"
)

func main() {
	// Setup Cosmos config
	sdkConfig := sdk.GetConfig()
	sdkConfig.SetBech32PrefixForAccount(
		app.Bech32MainPrefix,
		app.Bech32MainPrefix+sdk.PrefixPublic,
	)
	sdkConfig.SetBech32PrefixForValidator(
		app.Bech32MainPrefix+sdk.PrefixValidator+sdk.PrefixOperator,
		app.Bech32MainPrefix+sdk.PrefixValidator+sdk.PrefixOperator+sdk.PrefixPublic,
	)
	sdkConfig.SetBech32PrefixForConsensusNode(
		app.Bech32MainPrefix+sdk.PrefixValidator+sdk.PrefixConsensus,
		app.Bech32MainPrefix+sdk.PrefixValidator+sdk.PrefixConsensus+sdk.PrefixPublic,
	)
	sdkConfig.Seal()

	// Create Desmos codec
	codec := app.MakeCodec()

	rootCmd := &cobra.Command{
		Use:   "desmosp",
		Short: "Desmos Parser Command Line Interface",
	}

	rootCmd.AddCommand(
		version.GetVersionCmd(),
		parse.GetParseCmd(codec),
	)

	executor := PrepareMainCmd(rootCmd)
	err := executor.Execute()
	if err != nil {
		panic(err)
	}
}

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
