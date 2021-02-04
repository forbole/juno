package cmd

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/spf13/viper"

	"github.com/spf13/cobra"

	"github.com/desmos-labs/juno/config"
)

const (
	LogFormatJSON = "json"
	LogFormatText = "text"

	flagReplace = "replace"

	flagRPCAddress         = "rpc-address"
	flagGRPCAddress        = "grpc-address"
	flagGRPCInsecure       = "grpc-insecure"
	flagCosmosPrefix       = "cosmos-prefix"
	flagCosmosModules      = "cosmos-modules"
	flagDatabaseName       = "database-name"
	flagDatabaseHost       = "database-host"
	flagDatabasePort       = "database-port"
	flagDatabaseUser       = "database-user"
	flagDatabasePassword   = "database-password"
	flagDatabaseSSLMode    = "database-ssl-mode"
	flagDatabaseSchema     = "database-schema"
	flagLoggingLevel       = "logging-level"
	flagLoggingFormat      = "logging-format"
	flagParsingWorkers     = "parsing-workers"
	flagParsingNewBlocks   = "parsing-new-blocks"
	flagParsingOldBlocks   = "parsing-old-blocks"
	flagParsingStartHeight = "parsing-start-height"
	flagParsingFastSync    = "parsing-fast-sync"
)

// InitCmd returns the command that should be run in order to properly initialize BDJuno
func InitCmd(name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: fmt.Sprintf("Initializes the configuration files for %s", name),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {

			// Create the config path if not present
			folderPath := config.GetConfigFolderPath(name)
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
			configFilePath := config.GetConfigFilePath(name)
			file, _ := os.Stat(configFilePath)

			// Check if the file exists and replace is false
			if file != nil && !replace {
				return fmt.Errorf(
					"configuration file already present at %s. If you wish to overwrite it, use the --%s flag",
					configFilePath, flagReplace)
			}

			return config.Write(readConfigFromFlags(), configFilePath)
		},
	}

	cmd.Flags().Bool(flagReplace, false, "replaces any existing configuration with a new one")

	cmd.Flags().String(flagRPCAddress, "http://localhost:26657", "RPC address to use")

	cmd.Flags().String(flagGRPCAddress, "localhost:9090", "gRPC address to use")
	cmd.Flags().Bool(flagGRPCInsecure, true, "Tells whether the gRPC host should be treated as insecure or not")

	cmd.Flags().String(flagCosmosPrefix, "cosmos", "Bech32 prefix to use for addresses")
	cmd.Flags().StringSlice(flagCosmosModules, []string{}, "List of modules to use")

	cmd.Flags().String(flagDatabaseName, name, "Name of the database to use")
	cmd.Flags().String(flagDatabaseHost, "localhost", "Database host")
	cmd.Flags().Int64(flagDatabasePort, 5432, "Database port to use")
	cmd.Flags().String(flagDatabaseUser, "user", "User to use when authenticating inside the database")
	cmd.Flags().String(flagDatabasePassword, "password", "Password to use when authenticating inside the database")
	cmd.Flags().String(flagDatabaseSSLMode, "", "SSL mode to use when connecting to the database")
	cmd.Flags().String(flagDatabaseSchema, "public", "Database schema to use")

	cmd.Flags().String(flagLoggingLevel, zerolog.DebugLevel.String(), "Logging level to be used")
	cmd.Flags().String(flagLoggingFormat, LogFormatText, "Logging format to be used")

	cmd.Flags().Int64(flagParsingWorkers, 1, "Number of workers to use")
	cmd.Flags().Bool(flagParsingNewBlocks, true, "Whether or not to parse new blocks")
	cmd.Flags().Bool(flagParsingOldBlocks, true, "Whether or not to parse old blocks")
	cmd.Flags().Int64(flagParsingStartHeight, 1, "Starting height when parsing new blocks")
	cmd.Flags().Bool(flagParsingFastSync, true, "Whether to use fast sync or not when parsing old blocks")

	return cmd
}

func readConfigFromFlags() *config.Config {
	cfg := config.NewConfig(
		config.NewRPCConfig(
			viper.GetString(flagRPCAddress),
		),
		config.NewGrpcConfig(
			viper.GetString(flagGRPCAddress),
			viper.GetBool(flagGRPCInsecure),
		),
		config.NewCosmosConfig(
			viper.GetString(flagCosmosPrefix),
			viper.GetStringSlice(flagCosmosModules),
		),
		config.NewDatabaseConfig(
			viper.GetString(flagDatabaseName),
			viper.GetString(flagDatabaseHost),
			viper.GetInt64(flagDatabasePort),
			viper.GetString(flagDatabaseUser),
			viper.GetString(flagDatabasePassword),
			viper.GetString(flagDatabaseSSLMode),
			viper.GetString(flagDatabaseSchema),
		),
		config.NewLoggingConfig(
			viper.GetString(flagLoggingLevel),
			viper.GetString(flagLoggingFormat),
		),
		config.NewParsingConfig(
			viper.GetInt64(flagParsingWorkers),
			viper.GetBool(flagParsingNewBlocks),
			viper.GetBool(flagParsingOldBlocks),
			viper.GetInt64(flagParsingStartHeight),
			viper.GetBool(flagParsingFastSync),
		),
	)
	return cfg
}
