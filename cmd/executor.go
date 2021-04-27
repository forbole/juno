package cmd

import (
	"fmt"
	"os"

	"github.com/desmos-labs/juno/cmd/parse"

	"github.com/desmos-labs/juno/types"

	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/libs/cli"
)

// BuildDefaultExecutor allows to build an Executor containing a root command that
// has the provided name and description and the default version and parse sub-commands implementations.
//
// registrar will be used to register custom modules. Be sure to provide an implementation that returns all
// the modules that you want to use. If you don't want any custom module, use modules.EmptyRegistrar.
//
// setupCfg method will be used to customize the SDK configuration. If you don't want any customization
// you can use the config.DefaultConfigSetup variable.
//
// encodingConfigBuilder is used to provide a codec that will later be used to deserialize the
// transaction messages. Make sure you register all the types you need properly.
//
// dbBuilder is used to provide the database that will be used to save the data. If you don't have any
// particular need, you can use the Create variable to build a default database instance.
func BuildDefaultExecutor(parseConfig *parse.Config) cli.Executor {
	rootCmd := RootCmd(parseConfig.GetName())

	rootCmd.AddCommand(
		VersionCmd(),
		InitCmd(),
		parse.ParseCmd(parseConfig),
	)

	return PrepareRootCmd(rootCmd)
}

// RootCmd allows to build the default root command having the given name
func RootCmd(name string) *cobra.Command {
	return &cobra.Command{
		Use:   name,
		Short: fmt.Sprintf("%s is a Cosmos SDK-based chain data aggregator and exporter", name),
		Long: fmt.Sprintf(`A Cosmos chain data aggregator. It improves the chain's data accessibility
by providing an indexed database exposing aggregated resources and models such as blocks, validators, pre-commits, 
transactions, and various aspects of the governance module. 
%s is meant to run with a GraphQL layer on top so that it even further eases the ability for developers and
downstream clients to answer queries such as "What is the average gas cost of a block?" while also allowing
them to compose more aggregate and complex queries.`, name),
	}
}

// PrepareRootCmd is meant to prepare the given command binding all the viper flags
func PrepareRootCmd(cmd *cobra.Command) cli.Executor {
	cmd.PersistentPreRunE = types.ConcatCobraCmdFuncs(
		types.BindFlagsLoadViper,
		cmd.PersistentPreRunE,
	)
	return cli.Executor{Command: cmd, Exit: os.Exit}
}
