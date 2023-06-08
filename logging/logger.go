package logging

import (
	tmctypes "github.com/cometbft/cometbft/rpc/core/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/forbole/juno/v5/modules"
	"github.com/forbole/juno/v5/types"
)

const (
	LogKeyModule  = "module"
	LogKeyHeight  = "height"
	LogKeyTxHash  = "tx_hash"
	LogKeyMsgType = "msg_type"
)

// Logger defines a function that takes an error and logs it.
type Logger interface {
	SetLogLevel(level string) error
	SetLogFormat(format string) error

	Info(msg string, keyvals ...interface{})
	Debug(msg string, keyvals ...interface{})
	Error(msg string, keyvals ...interface{})

	GenesisError(module modules.Module, err error)
	BlockError(module modules.Module, block *tmctypes.ResultBlock, err error)
	EventsError(module modules.Module, results *tmctypes.ResultBlock, err error)
	TxError(module modules.Module, tx *types.Tx, err error)
	MsgError(module modules.Module, tx *types.Tx, msg sdk.Msg, err error)
}
