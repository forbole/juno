package init

import (
	"fmt"
	"os"

	"github.com/desmos-labs/juno/types"

	"github.com/spf13/cobra"
)

const (
	flagReplace = "replace"

	flagRPCClientName = "client-name"
	flagRPCAddress    = "rpc-address"

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
	flagGenesisFilePath     = "parsing-genesis-file-path"
	flagParsingStartHeight  = "parsing-start-height"
	flagParsingFastSync     = "parsing-fast-sync"

	flagPruningKeepRecent = "pruning-keep-recent"
	flagPruningKeepEvery  = "pruning-keep-every"
	flagPruningInterval   = "pruning-interval"

	flagTelemetryEnabled = "telemetry-enabled"
	flagTelemetryPort    = "telemetry-port"
)

// InitCmd returns the command that should be run in order to properly initialize BDJuno
func InitCmd(cfg *Config) *cobra.Command {
	command := &cobra.Command{
		Use:   "init",
		Short: "Initializes the configuration files",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create the config path if not present
			if _, err := os.Stat(types.HomePath); os.IsNotExist(err) {
				err = os.MkdirAll(types.HomePath, os.ModePerm)
				if err != nil {
					return err
				}
			}

			replace, err := cmd.Flags().GetBool(flagReplace)
			if err != nil {
				return err
			}

			// Get the config file
			configFilePath := types.GetConfigFilePath()
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

	defRPCConfig := types.DefaultRPCConfig()
	command.Flags().String(flagRPCClientName, defRPCConfig.GetClientName(), "Name of the subscriber to use when listening to events")
	command.Flags().String(flagRPCAddress, defRPCConfig.GetAddress(), "RPC address to use")

	defGRPCConfig := types.DefaultGrpcConfig()
	command.Flags().String(flagGRPCAddress, defGRPCConfig.GetAddress(), "gRPC address to use")
	command.Flags().Bool(flagGRPCInsecure, defGRPCConfig.IsInsecure(), "Tells whether the gRPC host should be treated as insecure or not")

	defCosmosConfig := types.DefaultCosmosConfig()
	command.Flags().String(flagCosmosPrefix, defCosmosConfig.GetPrefix(), "Bech32 prefix to use for addresses")
	command.Flags().StringSlice(flagCosmosModules, defCosmosConfig.GetModules(), "List of modules to use")

	defDatabaseConfig := types.DefaultDatabaseConfig()
	command.Flags().String(flagDatabaseName, defDatabaseConfig.GetName(), "Name of the database to use")
	command.Flags().String(flagDatabaseHost, defDatabaseConfig.GetHost(), "Database host")
	command.Flags().Int64(flagDatabasePort, defDatabaseConfig.GetPort(), "Database port to use")
	command.Flags().String(flagDatabaseUser, defDatabaseConfig.GetUser(), "User to use when authenticating inside the database")
	command.Flags().String(flagDatabasePassword, defDatabaseConfig.GetPassword(), "Password to use when authenticating inside the database")
	command.Flags().String(flagDatabaseSSLMode, defDatabaseConfig.GetSSLMode(), "SSL mode to use when connecting to the database")
	command.Flags().String(flagDatabaseSchema, defDatabaseConfig.GetSchema(), "Database schema to use")
	command.Flags().Int(flagDatabaseMaxOpenConnections, defDatabaseConfig.GetMaxOpenConnections(), "Max open connections (a value less than or equal to 0 means unlimited)")
	command.Flags().Int(flagDatabaseMaxIdleConnections, defDatabaseConfig.GetMaxIdleConnections(), "Max connections in the idle state (a value less than or equal to 0 means unlimited)")

	defLoggingConfig := types.DefaultLoggingConfig()
	command.Flags().String(flagLoggingLevel, defLoggingConfig.GetLogLevel(), "Logging level to be used")
	command.Flags().String(flagLoggingFormat, defLoggingConfig.GetLogFormat(), "Logging format to be used")

	defParsingConfig := types.DefaultParsingConfig()
	command.Flags().Int64(flagParsingWorkers, defParsingConfig.GetWorkers(), "Number of workers to use")
	command.Flags().Bool(flagParsingNewBlocks, defParsingConfig.ShouldParseNewBlocks(), "Whether or not to parse new blocks")
	command.Flags().Bool(flagParsingOldBlocks, defParsingConfig.ShouldParseOldBlocks(), "Whether or not to parse old blocks")
	command.Flags().Bool(flagParsingParseGenesis, defParsingConfig.ShouldParseGenesis(), "Whether or not to parse the genesis")
	command.Flags().String(flagGenesisFilePath, defParsingConfig.GetGenesisFilePath(), "(Optional) Path to the genesis file, if it should not be retrieved from the RPC")
	command.Flags().Int64(flagParsingStartHeight, defParsingConfig.GetStartHeight(), "Starting height when parsing new blocks")
	command.Flags().Bool(flagParsingFastSync, defParsingConfig.UseFastSync(), "Whether to use fast sync or not when parsing old blocks")

	defPruningConfig := types.DefaultPruningConfig()
	command.Flags().Int64(flagPruningKeepRecent, defPruningConfig.GetKeepRecent(), "Number of recent states to keep")
	command.Flags().Int64(flagPruningKeepEvery, defPruningConfig.GetKeepEvery(), "Keep every x amount of states forever")
	command.Flags().Int64(flagPruningInterval, defPruningConfig.GetInterval(), "Number of blocks every which to perform the pruning")

	defTelemetryConfig := types.DefaultTelemetryConfig()
	command.Flags().Bool(flagTelemetryEnabled, false, "Whether the telemetry server should be enabled or not")
	command.Flags().Uint(flagTelemetryPort, defTelemetryConfig.GetPort(), "Port on which the telemetry server will listen")

	// Set additional flags
	cfg.GetConfigSetupFlag()(command)

	return command
}
