package messages

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/forbole/juno/v3/database"
	"github.com/forbole/juno/v3/types"
)

// HandleMsg represents a message handler
func HandleMsg(
	index int, msg sdk.Msg, tx *types.Tx, cdc codec.Codec, db database.Database,
) error {

	return nil
}
