package init

import (
	"github.com/spf13/cobra"

	"github.com/desmos-labs/juno/types"
)

// ConfigFlagSetup represents a function that will be called in order to setup all the flags for the "init" command.
// Here you should add any flag that might be used from the user in order to set default configuration values when
// initializing for the first time the configuration of the command.
type ConfigFlagSetup = func(cmd *cobra.Command)

// DefaultFlagSetup represents a ConfigFlagSetup that sets no flag other than the default ones.
func DefaultFlagSetup(_ *cobra.Command) {}

// --------------------------------------------------------------------------------------------------------------------

// ConfigCreator represents a function that builds a Config instance from the flags that have been specified by the
// user inside the given command.
type ConfigCreator = func(cmd *cobra.Command) types.Config

// DefaultConfigCreator represents the default configuration creator that builds a Config instance using the values
// specified using the default flags.
func DefaultConfigCreator(cmd *cobra.Command) types.Config {
	rpcClientName, _ := cmd.Flags().GetString(flagRPCClientName)
	rpcAddr, _ := cmd.Flags().GetString(flagRPCAddress)

	grpcAddr, _ := cmd.Flags().GetString(flagGRPCAddress)
	grpcInsecure, _ := cmd.Flags().GetBool(flagGRPCInsecure)

	cosmosPrefix, _ := cmd.Flags().GetString(flagCosmosPrefix)
	cosmosModules, _ := cmd.Flags().GetStringSlice(flagCosmosModules)

	dbName, _ := cmd.Flags().GetString(flagDatabaseName)
	dbHost, _ := cmd.Flags().GetString(flagDatabaseHost)
	dbPort, _ := cmd.Flags().GetInt64(flagDatabasePort)
	dbUser, _ := cmd.Flags().GetString(flagDatabaseUser)
	dbPassword, _ := cmd.Flags().GetString(flagDatabasePassword)
	dbSSLMode, _ := cmd.Flags().GetString(flagDatabaseSSLMode)
	dbSchema, _ := cmd.Flags().GetString(flagDatabaseSchema)
	dbMaxOpenConnections, _ := cmd.Flags().GetInt(flagDatabaseMaxOpenConnections)
	dbMaxIdleConnections, _ := cmd.Flags().GetInt(flagDatabaseMaxIdleConnections)

	loggingLevel, _ := cmd.Flags().GetString(flagLoggingLevel)
	loggingFormat, _ := cmd.Flags().GetString(flagLoggingFormat)

	parsingWorkers, _ := cmd.Flags().GetInt64(flagParsingWorkers)
	parsingNewBlocks, _ := cmd.Flags().GetBool(flagParsingNewBlocks)
	parsingOldBlocks, _ := cmd.Flags().GetBool(flagParsingOldBlocks)
	parsingParseGenesis, _ := cmd.Flags().GetBool(flagParsingParseGenesis)
	parsingGenesisFilePath, _ := cmd.Flags().GetString(flagGenesisFilePath)
	parsingStartHeight, _ := cmd.Flags().GetInt64(flagParsingStartHeight)
	parsingFastSync, _ := cmd.Flags().GetBool(flagParsingFastSync)

	pruningKeepEvery, _ := cmd.Flags().GetInt64(flagPruningKeepEvery)
	pruningKeepRecent, _ := cmd.Flags().GetInt64(flagPruningKeepRecent)
	pruningInterval, _ := cmd.Flags().GetInt64(flagPruningInterval)

	return types.NewConfig(
		types.NewRPCConfig(rpcClientName, rpcAddr),
		types.NewGrpcConfig(grpcAddr, grpcInsecure),
		types.NewCosmosConfig(cosmosPrefix, cosmosModules),
		types.NewDatabaseConfig(
			dbName,
			dbHost,
			dbPort,
			dbUser,
			dbPassword,
			dbSSLMode,
			dbSchema,
			dbMaxOpenConnections,
			dbMaxIdleConnections,
		),
		types.NewLoggingConfig(loggingLevel, loggingFormat),
		types.NewParsingConfig(
			parsingWorkers,
			parsingNewBlocks,
			parsingOldBlocks,
			parsingParseGenesis,
			parsingGenesisFilePath,
			parsingStartHeight,
			parsingFastSync,
		),
		types.NewPruningConfig(
			pruningKeepRecent,
			pruningKeepEvery,
			pruningInterval,
		),
	)
}

// --------------------------------------------------------------------------------------------------------------------

// Config contains the configuration data for the init command
type Config struct {
	setupConfigFlag ConfigFlagSetup
	createConfig    ConfigCreator
}

// NewConfig allows to build a new Config instance
func NewConfig() *Config {
	return &Config{}
}

// WithConfigFlagSetup sets the given setup function as the flag setup
func (c *Config) WithConfigFlagSetup(setup ConfigFlagSetup) *Config {
	c.setupConfigFlag = setup
	return c
}

// GetConfigSetupFlag return the function that should be run to setup the flags that will later be used to build
// a default instance of the configuration object
func (c *Config) GetConfigSetupFlag() ConfigFlagSetup {
	if c.setupConfigFlag == nil {
		return DefaultFlagSetup
	}
	return c.setupConfigFlag
}

// WithConfigCreator sets the given setup function as the configuration creator
func (c *Config) WithConfigCreator(creator ConfigCreator) *Config {
	c.createConfig = creator
	return c
}

// GetConfigCreator return the function that should be run to create a configuration from a set of
// flags specified by the user with the "init" command
func (c *Config) GetConfigCreator() ConfigCreator {
	if c.setupConfigFlag == nil {
		return DefaultConfigCreator
	}
	return c.createConfig
}
