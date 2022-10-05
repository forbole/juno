package rawmessages

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/forbole/juno/v3/database"
	"github.com/forbole/juno/v3/types"
)

// HandleMsg represents a message handler that stores the given message inside the proper database table
func HandleMsg(
	index int, msg *codectypes.Any, tx *types.Tx, cdc codec.Codec, db database.Database,
) error {
	var addresses [][]string
	var involvedAddresses []string
	var rawMsg map[string]json.RawMessage

	// marshal msg value
	bz, err := cdc.MarshalJSON(msg)
	if err != nil {
		return err
	}

	// unmarshal the value properly
	err = json.Unmarshal(bz, &rawMsg)
	if err != nil {
		return err
	}

	msgValue, err := json.Marshal(&rawMsg)
	if err != nil {
		return err
	}

	// find all addresses contained inside the data string
	bech32AddrPrefix := sdk.GetConfig().GetBech32AccountAddrPrefix()
	addressRegex := regexp.MustCompile(fmt.Sprintf("%s[0-9a-zA-Z]+", bech32AddrPrefix))
	if addressRegex.MatchString(string(bz)) {
		addresses = addressRegex.FindAllStringSubmatch(string(bz), -1)
	}

	// rewrite into 1d array
	for _, addr := range addresses {
		involvedAddresses = append(involvedAddresses, addr...)
	}
	msgType := msg.TypeUrl[1:] // remove head "/"

	return db.SaveMessage(types.NewMessage(
		tx.TxHash,
		index,
		msgType,
		string(msgValue),
		involvedAddresses,
		tx.Height,
	))
}
