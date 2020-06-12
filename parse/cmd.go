package parse

import (
	"sync"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/desmos-labs/juno/config"
	"github.com/desmos-labs/juno/db"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

const (
	logLevelJSON = "json"
	logLevelText = "text"
)

var (
	waitGroup            sync.WaitGroup
	additionalOperations []AdditionalOperation
)

// SetupFlags allows to setup the given cmd by setting the required parse flags
func SetupFlags(cmd *cobra.Command) *cobra.Command {
	cmd.Flags().Int64(config.FlagStartHeight, 1, "sync missing or failed blocks starting from a given height")
	cmd.Flags().Int64(config.FlagWorkerCount, 1, "number of workers to run concurrently")
	cmd.Flags().Bool(config.FlagParseOldBlocks, true, "parse old blocks")
	cmd.Flags().Bool(config.FlagListenNewBlocks, true, "listen to new blocks")
	cmd.Flags().String(config.FlagLogLevel, zerolog.InfoLevel.String(), "logging level")
	cmd.Flags().String(config.FlagLogFormat, logLevelJSON, "logging format; must be either json or text")
	return cmd
}

// RegisterAdditionalOperation allows to register a new additional operation to be
// performed when the default setup has been completed successfully
func RegisterAdditionalOperation(operation AdditionalOperation) {
	additionalOperations = append(additionalOperations, operation)
}

// GetParseCmd returns the command that should be run when we want to start parsing a chain state
func GetParseCmd(cdc *codec.Codec, builder db.Builder) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "parse [config-file]",
		Short: "Start parsing a blockchain using the provided config file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return ParseCmdHandler(cdc, builder, args[0], additionalOperations)
		},
	}

	return SetupFlags(cmd)
}
