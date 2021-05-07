package init

import (
	"fmt"
	"os"

	"github.com/desmos-labs/juno/types"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

const (
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
func InitCmd(name string, cfg *Config) *cobra.Command {
	command := &cobra.Command{
		Use:   "init",
		Short: "Initializes the configuration files",
		RunE: func(cmd *cobra.Command, args []string) error {
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

			// Get the config from the flags
			config := cfg.GetConfigCreator()(cmd)
			return types.Write(config, configFilePath)
		},
	}

	// Set default flags
	command.Flags().Bool(flagReplace, false, "replaces any existing configuration with a new one")

	command.Flags().String(flagRPCAddress, "http://localhost:26657", "RPC address to use")

	command.Flags().String(flagGRPCAddress, "localhost:9090", "gRPC address to use")
	command.Flags().Bool(flagGRPCInsecure, true, "Tells whether the gRPC host should be treated as insecure or not")

	command.Flags().String(flagCosmosPrefix, "cosmos", "Bech32 prefix to use for addresses")
	command.Flags().StringSlice(flagCosmosModules, []string{}, "List of modules to use")

	command.Flags().String(flagDatabaseName, "database-name", "Name of the database to use")
	command.Flags().String(flagDatabaseHost, "localhost", "Database host")
	command.Flags().Int64(flagDatabasePort, 5432, "Database port to use")
	command.Flags().String(flagDatabaseUser, "user", "User to use when authenticating inside the database")
	command.Flags().String(flagDatabasePassword, "password", "Password to use when authenticating inside the database")
	command.Flags().String(flagDatabaseSSLMode, "", "SSL mode to use when connecting to the database")
	command.Flags().String(flagDatabaseSchema, "public", "Database schema to use")
	command.Flags().Int(flagDatabaseMaxOpenConnections, 0, "Max open connections (a value less than or equal to 0 means unlimited)")
	command.Flags().Int(flagDatabaseMaxIdleConnections, 0, "Max connections in the idle state (a value less than or equal to 0 means unlimited)")

	command.Flags().String(flagLoggingLevel, zerolog.DebugLevel.String(), "Logging level to be used")
	command.Flags().String(flagLoggingFormat, LogFormatText, "Logging format to be used")

	command.Flags().Int64(flagParsingWorkers, 1, "Number of workers to use")
	command.Flags().Bool(flagParsingNewBlocks, true, "Whether or not to parse new blocks")
	command.Flags().Bool(flagParsingOldBlocks, true, "Whether or not to parse old blocks")
	command.Flags().Bool(flagParsingParseGenesis, true, "Whether or not to parse the genesis")
	command.Flags().Int64(flagParsingStartHeight, 1, "Starting height when parsing new blocks")
	command.Flags().Bool(flagParsingFastSync, true, "Whether to use fast sync or not when parsing old blocks")

	// Set additional flags
	cfg.GetConfigSetupFlag()(command)

	return command
}
