package logging

import (
	tmctypes "github.com/cometbft/cometbft/rpc/core/types"

	"github.com/0xPellNetwork/juno/v6/modules"
	"github.com/0xPellNetwork/juno/v6/types"
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
	TxError(module modules.Module, tx *types.Transaction, err error)
	MsgError(module modules.Module, tx *types.Transaction, msg types.Message, err error)
}
