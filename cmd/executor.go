package cmd

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/desmos-labs/juno/config"
	"github.com/desmos-labs/juno/types"

	"github.com/desmos-labs/juno/modules/registrar"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/cli"

	"github.com/desmos-labs/juno/db"
)

// BuildDefaultExecutor allows to build an Executor containing a root command that
// has the provided name and description and the default version and parse sub-commands implementations.
//
// registrar will be used to register custom modules. Be sure to provide an implementation that returns all
// the modules that you want to use. If you don't want any custom module, use modules.EmptyRegistrar.
//
// setupCfg method will be used to customize the SDK configuration. If you don't want any customization
// you can use the config.DefaultSetup variable.
//
// encodingConfigBuilder is used to provide a codec that will later be used to deserialize the
// transaction messages. Make sure you register all the types you need properly.
//
// dbBuilder is used to provide the database that will be used to save the data. If you don't have any
// particular need, you can use the Create variable to build a default database instance.
func BuildDefaultExecutor(
	name string, registrar registrar.Registrar,
	setupCfg types.SdkConfigSetup, encodingConfigBuilder types.EncodingConfigBuilder, dbBuilder db.Builder,
) cli.Executor {
	rootCmd := RootCmd(name)

	rootCmd.AddCommand(
		VersionCmd(),
		InitCmd(),
		ParseCmd(name, registrar, encodingConfigBuilder, setupCfg, dbBuilder),
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
	cmd.PersistentPreRunE = concatCobraCmdFuncs(
		bindFlagsLoadViper,
		cmd.PersistentPreRunE,
	)
	return cli.Executor{Command: cmd, Exit: os.Exit}
}

// cobraCmdFunc represents a cobra command function
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

// Bind all flags and read the config into viper
func bindFlagsLoadViper(cmd *cobra.Command, _ []string) error {
	// cmd.Flags() includes flags from this command and all persistent flags from the parent
	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		return err
	}

	return nil
}

// readConfig parses the configuration file
func readConfig(name string) cobraCmdFunc {
	return func(_ *cobra.Command, _ []string) error {
		file := config.GetConfigFilePath(name)

		// Make sure the path exists
		if _, err := os.Stat(file); os.IsNotExist(err) {
			return fmt.Errorf("%s file does not exist. Make sure you have run %s init", file, name)
		}

		cfg, err := config.Read(file)
		if err != nil {
			return err
		}

		// Set the global configuration
		config.Cfg = cfg
		return nil
	}
}

// setupLogging setups the logging for the entire project
func setupLogging(_ *cobra.Command, _ []string) error {
	// Init logging level
	logLvl, err := zerolog.ParseLevel(config.Cfg.Logging.LogLevel)
	if err != nil {
		return err
	}
	zerolog.SetGlobalLevel(logLvl)

	// Init logging format
	switch config.Cfg.Logging.LogFormat {
	case LogFormatJSON:
		// JSON is the default logging format
		break

	case LogFormatText:
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
		break

	default:
		return fmt.Errorf("invalid logging format: %s", config.Cfg.Logging.LogFormat)
	}
	return err
}
