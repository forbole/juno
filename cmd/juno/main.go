package main

import (
	"fmt"

	"github.com/angelorc/desmos-parser/config"
	"github.com/angelorc/desmos-parser/db"
	"github.com/angelorc/desmos-parser/db/mongo"
	"github.com/angelorc/desmos-parser/parse"
	"github.com/angelorc/desmos-parser/parse/worker"
	"github.com/angelorc/desmos-parser/types"
	"github.com/angelorc/desmos-parser/version"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/desmos-labs/desmos/app"
	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/libs/cli"
)

func main() {
	// Register custom handlers
	worker.RegisterMsgHandler(msgHandler)

	// Build the executor
	executor := BuildExecutor("juno", setupConfig, app.MakeCodec, mongo.Builder)

	// Run the commands and panic on any error
	err := executor.Execute()
	if err != nil {
		panic(err)
	}
}

// Utility method that allows to build an Executor containing a root command that
// has the provided name and description.
//
// The provided setupCfg method will be used to customize the SDK configuration. If you don't want any customization
// you can use the types.EmptySetup method.
//
// The provided cdcBuilder is used to provide a codec that will later be used to deserialize the
// transaction messages. Make sure you register all the types you need properly.
//
// The provided dbBuilder is used to provide the database that will be used to save the data.
func BuildExecutor(name string, setupCfg types.SdkConfigSetup, cdcBuilder types.CodecBuilder, dbBuilder db.Builder) cli.Executor {

	sdkConfig := sdk.GetConfig()
	setupCfg(sdkConfig)
	sdkConfig.Seal()

	rootCmd := &cobra.Command{
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

	rootCmd.AddCommand(
		version.GetVersionCmd(),
		parse.GetParseCmd(cdcBuilder(), dbBuilder),
	)

	return config.PrepareMainCmd(rootCmd)
}

// TODO: Move this inside the Desmos implementation
func setupConfig(cfg *sdk.Config) {
	cfg.SetBech32PrefixForAccount(
		app.Bech32MainPrefix,
		app.Bech32MainPrefix+sdk.PrefixPublic,
	)
	cfg.SetBech32PrefixForValidator(
		app.Bech32MainPrefix+sdk.PrefixValidator+sdk.PrefixOperator,
		app.Bech32MainPrefix+sdk.PrefixValidator+sdk.PrefixOperator+sdk.PrefixPublic,
	)
	cfg.SetBech32PrefixForConsensusNode(
		app.Bech32MainPrefix+sdk.PrefixValidator+sdk.PrefixConsensus,
		app.Bech32MainPrefix+sdk.PrefixValidator+sdk.PrefixConsensus+sdk.PrefixPublic,
	)
}
