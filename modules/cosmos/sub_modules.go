package cosmos

import (
	"encoding/json"

	tmtypes "github.com/cometbft/cometbft/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	"github.com/forbole/juno/v5/types"
)

type GenesisModule interface {
	// HandleGenesis allows to handle the genesis state.
	// For convenience of use, the already-unmarshalled AppState is provided along with the full GenesisDoc.
	// NOTE. The returned error will be logged using the GenesisError method. All other modules' handlers
	// will still be called.
	HandleGenesis(doc *tmtypes.GenesisDoc, appState map[string]json.RawMessage) error
}

type TransactionModule interface {
	// HandleTx handles a single transaction.
	// For each message present inside the transaction, HandleMsg will be called as well.
	// NOTE. The returned error will be logged using the TxError method. All other modules' handlers
	// will still be called.
	HandleTx(tx *types.Tx) error
}

type MessageModule interface {
	// HandleMsg handles a single message.
	// For convenience of use, the index of the message inside the transaction and the transaction itself
	// are passed as well.
	// NOTE. The returned error will be logged using the MsgError method. All other modules' handlers
	// will still be called.
	HandleMsg(index int, msg sdk.Msg, tx *types.Tx) error
}

type AuthzMessageModule interface {
	// HandleMsgExec handles a single message that is contained within an authz.MsgExec instance.
	// For convenience of use, the index of the message inside the transaction and the transaction itself
	// are passed as well.
	// NOTE. The returned error will be logged using the MsgError method. All other modules' handlers
	// will still be called.
	HandleMsgExec(index int, msgExec *authz.MsgExec, authzMsgIndex int, executedMsg sdk.Msg, tx *types.Tx) error
}
