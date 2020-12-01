package cmd

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/desmos-labs/juno/config"
	"github.com/desmos-labs/juno/db"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/cli"
	"os"
)

// BuildDefaultExecutor allows to build an Executor containing a root command that
// has the provided name and description and the default version and parse sub-commands implementations.
//
// The provided setupCfg method will be used to customize the SDK configuration. If you don't want any customization
// you can use the config.EmptySetup variable.
//
// The provided cdcBuilder is used to provide a codec that will later be used to deserialize the
// transaction messages. Make sure you register all the types you need properly.
//
// The provided dbBuilder is used to provide the database that will be used to save the data. If you don't have any
// particular need, you can use the Create variable to build a default database instance.
func BuildDefaultExecutor(
	name string, setupCfg config.SdkConfigSetup, cdcBuilder config.CodecBuilder, dbBuilder db.Builder,
) cli.Executor {
	// Build the codec
	cdc := cdcBuilder()

	// Setup the SDK configuration
	sdkConfig := sdk.GetConfig()
	setupCfg(sdkConfig)
	sdkConfig.Seal()

	rootCmd := &cobra.Command{
		Use:   name,
		Short: fmt.Sprintf("%s is a Cosmos SDK-based chain data aggregator and exporter", name),
		Long: fmt.Sprintf(`A Cosmos SDK-based chain data aggregator. It improves the chain's data accessibility
by providing an indexed database exposing aggregated resources and models such as blocks, validators, pre-commits, 
transactions, and various aspects of the governance module. 
%s is meant to run with a GraphQL layer on top so that it even further eases the ability for developers and
downstream clients to answer queries such as "What is the average gas cost of a block?" while also allowing
them to compose more aggregate and complex queries.`, name),
	}

	rootCmd.AddCommand(
		VersionCmd(),
		ParseCmd(cdc, dbBuilder),
	)

	return PrepareMainCmd(rootCmd)
}

const (
	FlagStartHeight     = "start-height"
	FlagWorkerCount     = "workers"
	FlagListenNewBlocks = "listen-new-blocks"
	FlagParseOldBlocks  = "parse-old-blocks"
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
