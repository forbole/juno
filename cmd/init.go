package cmd

import (
	"fmt"
	"os"

	"github.com/desmos-labs/juno/types"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

const (
	LogFormatJSON = "json"
	LogFormatText = "text"

	flagReplace = "replace"

	flagRPCAddress   = "rpc-address"
	flagGRPCAddress  = "grpc-address"
	flagGRPCInsecure = "grpc-insecure"

	flagCosmosPrefix  = "cosmos-prefix"
	flagCosmosModules = "cosmos-modules"

	flagDatabaseName               = "database-name"
	flagDatabaseHost               = "database-host"
	flagDatabasePort               = "database-port"
	flagDatabaseUser               = "database-user"
	flagDatabasePassword           = "database-password"
	flagDatabaseSSLMode            = "database-ssl-mode"
	flagDatabaseSchema             = "database-schema"
	flagDatabaseMaxOpenConnections = "max-open-connections"
	flagDatabaseMaxIdleConnections = "max-idle-connections"

	flagLoggingLevel  = "logging-level"
	flagLoggingFormat = "logging-format"

	flagParsingWorkers      = "parsing-workers"
	flagParsingNewBlocks    = "parsing-new-blocks"
	flagParsingOldBlocks    = "parsing-old-blocks"
	flagParsingParseGenesis = "parsing-parse-genesis"
	flagParsingStartHeight  = "parsing-start-height"
	flagParsingFastSync     = "parsing-fast-sync"
)

// InitCmd returns the command that should be run in order to properly initialize BDJuno
func InitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initializes the configuration files",
		RunE: func(cmd *cobra.Command, args []string) error {
			name := cmd.Root().Name()

			// Create the config path if not present
			folderPath := types.GetConfigFolderPath(name)
			if _, err := os.Stat(folderPath); os.IsNotExist(err) {
				err = os.MkdirAll(folderPath, os.ModePerm)
				if err != nil {
					return err
				}
			}

			replace, err := cmd.Flags().GetBool(flagReplace)
			if err != nil {
				return err
			}

			// Get the config file
			configFilePath := types.GetConfigFilePath(name)
			file, _ := os.Stat(configFilePath)

			// Check if the file exists and replace is false
			if file != nil && !replace {
				return fmt.Errorf(
					"configuration file already present at %s. If you wish to overwrite it, use the --%s flag",
					configFilePath, flagReplace)
			}

			return types.Write(readConfigFromFlags(cmd), configFilePath)
		},
	}

	cmd.Flags().Bool(flagReplace, false, "replaces any existing configuration with a new one")

	cmd.Flags().String(flagRPCAddress, "http://localhost:26657", "RPC address to use")

	cmd.Flags().String(flagGRPCAddress, "localhost:9090", "gRPC address to use")
	cmd.Flags().Bool(flagGRPCInsecure, true, "Tells whether the gRPC host should be treated as insecure or not")

	cmd.Flags().String(flagCosmosPrefix, "cosmos", "Bech32 prefix to use for addresses")
	cmd.Flags().StringSlice(flagCosmosModules, []string{}, "List of modules to use")

	cmd.Flags().String(flagDatabaseName, "database-name", "Name of the database to use")
	cmd.Flags().String(flagDatabaseHost, "localhost", "Database host")
	cmd.Flags().Int64(flagDatabasePort, 5432, "Database port to use")
	cmd.Flags().String(flagDatabaseUser, "user", "User to use when authenticating inside the database")
	cmd.Flags().String(flagDatabasePassword, "password", "Password to use when authenticating inside the database")
	cmd.Flags().String(flagDatabaseSSLMode, "", "SSL mode to use when connecting to the database")
	cmd.Flags().String(flagDatabaseSchema, "public", "Database schema to use")
	cmd.Flags().Int(flagDatabaseMaxOpenConnections, 0, "Max open connections (a value less than or equal to 0 means unlimited)")
	cmd.Flags().Int(flagDatabaseMaxIdleConnections, 0, "Max connections in the idle state (a value less than or equal to 0 means unlimited)")

	cmd.Flags().String(flagLoggingLevel, zerolog.DebugLevel.String(), "Logging level to be used")
	cmd.Flags().String(flagLoggingFormat, LogFormatText, "Logging format to be used")

	cmd.Flags().Int64(flagParsingWorkers, 1, "Number of workers to use")
	cmd.Flags().Bool(flagParsingNewBlocks, true, "Whether or not to parse new blocks")
	cmd.Flags().Bool(flagParsingOldBlocks, true, "Whether or not to parse old blocks")
	cmd.Flags().Bool(flagParsingParseGenesis, true, "Whether or not to parse the genesis")
	cmd.Flags().Int64(flagParsingStartHeight, 1, "Starting height when parsing new blocks")
	cmd.Flags().Bool(flagParsingFastSync, true, "Whether to use fast sync or not when parsing old blocks")

	return cmd
}

func readConfigFromFlags(cmd *cobra.Command) types.Config {
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
