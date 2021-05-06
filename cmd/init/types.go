package init

import (
	"github.com/spf13/cobra"

	"github.com/desmos-labs/juno/types"
)

type FlagSetup = func(cmd *cobra.Command)

func DefaultFlagSetup(_ *cobra.Command) {}

// --------------------------------------------------------------------------------------------------------------------

type ConfigCreator = func(cmd *cobra.Command) types.Config

func DefaultConfigCreator(cmd *cobra.Command) types.Config {
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
	parsingStartHeight, _ := cmd.Flags().GetInt64(flagParsingStartHeight)
	parsingFastSync, _ := cmd.Flags().GetBool(flagParsingFastSync)

	return types.NewConfig(
		types.NewRPCConfig(rpcAddr),
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
			parsingStartHeight,
			parsingFastSync,
		),
	)
}

// --------------------------------------------------------------------------------------------------------------------

// Config contains the configuration data for the init command
type Config struct {
	setupFlag    FlagSetup
	createConfig ConfigCreator
}

// NewConfig allows to build a new Config instance
func NewConfig() *Config {
	return &Config{}
}

// WithFlagSetup sets the given setup function as the flag setup
func (c *Config) WithFlagSetup(setup FlagSetup) *Config {
	c.setupFlag = setup
	return c
}

// GetSetupFlag return the function that should be run to setup the flags
func (c *Config) GetSetupFlag() FlagSetup {
	if c.setupFlag == nil {
		return DefaultFlagSetup
	}
	return c.setupFlag
}

// WithConfigCreator sets the given setup function as the configuration creator
func (c *Config) WithConfigCreator(creator ConfigCreator) *Config {
	c.createConfig = creator
	return c
}

// GetConfigCreator return the function that should be run to create a configuration
func (c *Config) GetConfigCreator() ConfigCreator {
	if c.setupFlag == nil {
		return DefaultConfigCreator
	}
	return c.createConfig
}
