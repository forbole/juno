package rawmessages

import (
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

	var msgData sdk.MsgData
	var addresses [][]string
	var involvedAddresses []string

	// Unmarshal the value properly
	err := cdc.Unmarshal(msg.Value, &msgData)
	if err != nil {
		return fmt.Errorf("error when unmarshaling msg %s", err)
	}

	bech32AddrPrefix := sdk.GetConfig().GetBech32AccountAddrPrefix()
	addressRegex := regexp.MustCompile(fmt.Sprintf("%s[0-9a-zA-Z]+", bech32AddrPrefix))
	if addressRegex.MatchString(string(msgData.Data)) {
		addresses = addressRegex.FindAllStringSubmatch(string(msgData.Data), -1)
	}

	// rewrite into 1d array
	for _, addr := range addresses {
		involvedAddresses = append(involvedAddresses, addr...)
	}

	msgType := msgData.MsgType[1:] // remove head "/"

	return db.SaveMessage(types.NewMessage(
		tx.TxHash,
		index,
		msgType,
		msgData,
		involvedAddresses,
		tx.Height,
	))
}
