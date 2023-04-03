package messages

import (
	"encoding/json"
	"fmt"
	"regexp"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/forbole/juno/v4/types"
)

// HandleRawMsg implements modules.RawMessageModule
func (m *Module) HandleRawMsg(index int, msg *codectypes.Any, tx *types.Tx) error {
	// Get the msg value
	msgValueBz, err := m.parseMsgValue(msg)
	if err != nil {
		return err
	}
	msgValueJSON := string(msgValueBz)

	// Find all addresses contained inside the data string
	bech32AddrPrefix := sdk.GetConfig().GetBech32AccountAddrPrefix()
	addressRegex := regexp.MustCompile(fmt.Sprintf("%s[0-9a-zA-Z]{6,}", bech32AddrPrefix))
	involvedAddresses := addressRegex.FindAllString(string(msg.Value), -1)

	// Remove the leading "/"
	msgType := msg.TypeUrl[1:]

	return m.db.SaveMessage(types.NewMessage(
		tx.TxHash,
		index,
		msgType,
		msgValueJSON,
		involvedAddresses,
		tx.Height,
	))
}

// parseMsgValue reads the given codectypes.Any message and gets its inner value by serializing
// it to a JSON map and removing the msg_type key
func (m *Module) parseMsgValue(msg *codectypes.Any) ([]byte, error) {
	msgData := sdk.MsgData{
		MsgType: msg.TypeUrl,
		Data:    msg.Value,
	}

	msgJSONBz, err := m.cdc.MarshalJSON(&msgData)
	if err != nil {
		return nil, fmt.Errorf("error while marshalling msg any to json: %s", err)
	}
	var msgMap map[string]json.RawMessage
	err = json.Unmarshal(msgJSONBz, &msgMap)
	if err != nil {
		return nil, fmt.Errorf("error while unmarshalling msg to map: %s", err)
	}
	delete(msgMap, "msg_type")

	// Re-serialize the map without the msg_type key
	noTypeMsgBz, err := json.Marshal(&msgMap)
	if err != nil {
		return nil, fmt.Errorf("error while marshalling no type msg to json: %s", err)
	}
	return noTypeMsgBz, nil
}
