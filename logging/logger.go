package logging

import (
	bftcoretypes "github.com/cometbft/cometbft/rpc/core/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/forbole/juno/v4/modules"
	"github.com/forbole/juno/v4/types"
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
	BlockError(module modules.Module, block *bftcoretypes.ResultBlock, err error)
	EventsError(module modules.Module, results *bftcoretypes.ResultBlock, err error)
	TxError(module modules.Module, tx *types.Tx, err error)
	MsgError(module modules.Module, tx *types.Tx, msg sdk.Msg, err error)
	RawMsgError(module modules.Module, tx *types.Tx, msg *codectypes.Any, err error)
}
