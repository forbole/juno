package cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/forbole/juno/v4/types/config"

	initcmd "github.com/forbole/juno/v4/cmd/init"
	migratecmd "github.com/forbole/juno/v4/cmd/migrate"
	parsecmd "github.com/forbole/juno/v4/cmd/parse"
	startcmd "github.com/forbole/juno/v4/cmd/start"

	"github.com/forbole/juno/v4/types"

	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/libs/cli"
)

var (
	FlagHome = "home"
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
func BuildDefaultExecutor(config *Config) cli.Executor {
	rootCmd := RootCmd(config.GetName())

	rootCmd.AddCommand(
		VersionCmd(),
		initcmd.NewInitCmd(config.GetInitConfig()),
		parsecmd.NewParseCmd(config.GetParseConfig()),
		startcmd.NewStartCmd(config.GetParseConfig()),
		migratecmd.NewMigrateCmd(config.GetName(), config.GetParseConfig()),
	)

	return PrepareRootCmd(config.GetName(), rootCmd)
}

// RootCmd allows to build the default root command having the given name
func RootCmd(name string) *cobra.Command {
	return &cobra.Command{
		Use:   name,
		Short: fmt.Sprintf("%s is a Chain SDK-based chain data aggregator and exporter", name),
		Long: fmt.Sprintf(`A Chain chain data aggregator. It improves the chain's data accessibility
by providing an indexed database exposing aggregated resources and models such as blocks, validators, pre-commits, 
transactions, and various aspects of the governance module. 
%s is meant to run with a GraphQL layer on top so that it even further eases the ability for developers and
downstream clients to answer queries such as "What is the average gas cost of a block?" while also allowing
them to compose more aggregate and complex queries.`, name),
	}
}

// PrepareRootCmd is meant to prepare the given command binding all the viper flags
func PrepareRootCmd(name string, cmd *cobra.Command) cli.Executor {
	cmd.PersistentPreRunE = types.ConcatCobraCmdFuncs(
		types.BindFlagsLoadViper,
		setupHome,
		cmd.PersistentPreRunE,
	)

	home, _ := os.UserHomeDir()
	defaultConfigPath := path.Join(home, fmt.Sprintf(".%s", name))
	cmd.PersistentFlags().String(FlagHome, defaultConfigPath, "Set the home folder of the application, where all files will be stored")

	return cli.Executor{Command: cmd, Exit: os.Exit}
}

// setupHome setups the home directory of the root command
func setupHome(cmd *cobra.Command, _ []string) error {
	config.HomePath, _ = cmd.Flags().GetString(FlagHome)
	return nil
}
