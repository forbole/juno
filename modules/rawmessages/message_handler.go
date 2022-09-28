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

	// unmarshal the value properly
	err := cdc.Unmarshal(msg.Value, &msgData)
	if err != nil {
		return fmt.Errorf("error when unmarshaling msg %s", err)
	}

	// find all addresses contained inside the data string
	bech32AddrPrefix := sdk.GetConfig().GetBech32AccountAddrPrefix()
	addressRegex := regexp.MustCompile(fmt.Sprintf("%s[0-9a-zA-Z]+", bech32AddrPrefix))
	if addressRegex.MatchString(string(msgData.Data)) {
		addresses = addressRegex.FindAllStringSubmatch(string(msgData.Data), -1)
	}

	// rewrite into 1d array
	for _, addr := range addresses {
		involvedAddresses = append(involvedAddresses, addr...)
	}

	// marshal msgData value
	bz, err := cdc.MarshalJSON(&msgData)
	if err != nil {
		return err
	}

	fmt.Printf("\n bz %s \n", string(bz))
	fmt.Printf("\n msg %s \n", msg.GoString())

	msgType := msgData.MsgType[1:] // remove head "/"

	return db.SaveMessage(types.NewMessage(
		tx.TxHash,
		index,
		msgType,
		string(bz),
		involvedAddresses,
		tx.Height,
	))
}
