package logging

import (
	"fmt"
	"os"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	tmctypes "github.com/tendermint/tendermint/rpc/core/types"

	"github.com/desmos-labs/juno/modules"
	"github.com/desmos-labs/juno/types"
)

var _ Logger = &defaultLogger{}

// defaultLogger represents the default logger for any kind of error
type defaultLogger struct{}

// SetLogLevel implements Logger
func (d *defaultLogger) SetLogLevel(level string) error {
	logLvl, err := zerolog.ParseLevel(level)
	if err != nil {
		return err
	}

	zerolog.SetGlobalLevel(logLvl)
	return nil
}

// SetLogFormat implements Logger
func (d *defaultLogger) SetLogFormat(format string) error {
	switch format {
	case "json":
		// JSON is the default logging format
		break

	case "text":
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
		break

	default:
		return fmt.Errorf("invalid logging format: %s", format)
	}

	return nil
}

// LogGenesisError implements Logger
func (d *defaultLogger) LogGenesisError(module modules.Module, err error) {
	log.Error().Err(err).Str(LogKeyModule, module.Name()).
		Msg("error while handling genesis")
}

// LogBLockError implements Logger
func (d *defaultLogger) LogBLockError(module modules.Module, block *tmctypes.ResultBlock, err error) {
	log.Error().Err(err).Str(LogKeyModule, module.Name()).Int64(LogKeyHeight, block.Block.Height).
		Msg("error while handling block")
}

// LogTxError implements Logger
func (d *defaultLogger) LogTxError(module modules.Module, tx *types.Tx, err error) {
	log.Error().Err(err).Str(LogKeyModule, module.Name()).Int64("height", tx.Height).
		Str(LogKeyTxHash, tx.TxHash).Msg("error while handling transaction")
}

// LogMsgError implements Logger
func (d *defaultLogger) LogMsgError(module modules.Module, tx *types.Tx, msg sdk.Msg, err error) {
	log.Error().Err(err).Str(LogKeyModule, module.Name()).Int64(LogKeyHeight, tx.Height).
		Str(LogKeyTxHash, tx.TxHash).Str(LogKeyMsgType, msg.Type()).Msg("error while handling message")
}
