package x

import (
	"encoding/json"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/desmos-labs/juno/client"
	"github.com/desmos-labs/juno/db"
	"github.com/desmos-labs/juno/types"
	"github.com/go-co-op/gocron"
	tmctypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

var ModuleBasic []Module

// Module represents a generic module that has the ability to properly handle the chain data.
type Module interface {
	// Name returns the module name
	Name() string

	// RegisterPeriodicOperations allows to register all the operations that will be run on a periodic basis.
	// The given scheduler can be used to define the periodicity of each task.
	// NOTE. This method will only be run ONCE during the module initialization.
	RegisterPeriodicOperations(scheduler *gocron.Scheduler) error

	// RunAdditionalOperations runs all the additional operations required to the module.
	// This is the perfect place where to initialize all the operations that subscribe to websockets or other
	// external sources.
	// NOTE. This method will only be run ONCE before starting the parsing of the blocks.
	RunAdditionalOperations(cdc *codec.Codec, cp *client.Proxy, db db.Database) error

	// HandleGenesis allows to handle the genesis state.
	// For convenience of use, the already-unmarshalled AppState is provided along with the full GenesisDoc.
	// NOTE. The returned error will be logged using the logging.LogGenesisError method. All other modules' handlers
	// will still be called.
	HandleGenesis(
		doc *tmtypes.GenesisDoc, appState map[string]json.RawMessage,
		cdc *codec.Codec, cp *client.Proxy, db db.Database,
	) error

	// HandleBlock allows to handle a single block.
	// For convenience of use, all the transactions present inside the given block
	// and the currently used database will be passed as well.
	// For each transaction present inside the block, HandleTx will be called as well.
	// NOTE. The returned error will be logged using the logging.LogBlockError method. All other modules' handlers
	// will still be called.
	HandleBlock(
		block *tmctypes.ResultBlock, txs []types.Tx, vals *tmctypes.ResultValidators,
		cdc *codec.Codec, cp *client.Proxy, db db.Database,
	) error

	// HandleTx handles a single transaction.
	// For each message present inside the transaction, HandleMsg will be called as well.
	// NOTE. The returned error will be logged using the logging.LogTxError method. All other modules' handlers
	// will still be called.
	HandleTx(tx types.Tx, cdc *codec.Codec, cp *client.Proxy, db db.Database) error

	// HandleTx handles a single transaction.
	// For convenience of usa, the index of the message inside the transaction and the transaction itself
	// are passed as well.
	// NOTE. The returned error will be logged using the logging.LogMsgError method. All other modules' handlers
	// will still be called.
	HandleMsg(index int, msg sdk.Msg, tx types.Tx, cdc *codec.Codec, cp *client.Proxy, db db.Database) error
}
