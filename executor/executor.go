package executor

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/desmos-labs/juno/config"
	"github.com/desmos-labs/juno/db"
	"github.com/desmos-labs/juno/parse"
	"github.com/desmos-labs/juno/types"
	"github.com/desmos-labs/juno/version"
	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/libs/cli"
)

// BuildRootCmd allows to build the default root command having the specified name.
//
// The given SdkConfigSetup method will be used to setup the Cosmos SDK configuration before
// creating the command.
func BuildRootCmd(name string, setupCfg types.SdkConfigSetup) *cobra.Command {
	sdkConfig := sdk.GetConfig()
	setupCfg(sdkConfig)
	sdkConfig.Seal()

	return &cobra.Command{
		Use:   name,
		Short: fmt.Sprintf("%s is a Cosmos Hub data aggregator and exporter", name),
		Long: fmt.Sprintf(`A cosmos Hub data aggregator. It improves the Hub's data accessibility
by providing an indexed database exposing aggregated resources and
models such as blocks, validators, pre-commits, transactions, and various aspects
of the governance module. %s is meant to run with a GraphQL layer on top so that
it even further eases the ability for developers and downstream clients to answer
queries such as "what is the average gas cost of a block?" while also allowing
them to compose more aggregate and complex queries.`, name),
	}
}

// BuildDefaultExecutor allows to build an Executor containing a root command that
// has the provided name and description and the default version and parse sub-commands implementations.
//
// The provided setupCfg method will be used to customize the SDK configuration. If you don't want any customization
// you can use the types.EmptySetup method.
//
// The provided cdcBuilder is used to provide a codec that will later be used to deserialize the
// transaction messages. Make sure you register all the types you need properly.
//
// The provided dbBuilder is used to provide the database that will be used to save the data.
func BuildDefaultExecutor(name string, setupCfg types.SdkConfigSetup, cdcBuilder types.CodecBuilder, dbBuilder db.Builder) cli.Executor {
	rootCmd := BuildRootCmd(name, setupCfg)

	rootCmd.AddCommand(
		version.GetVersionCmd(),
		parse.GetParseCmd(cdcBuilder(), dbBuilder),
	)

	return config.PrepareMainCmd(rootCmd)
}
