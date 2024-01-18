package messages

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	"github.com/forbole/juno/v5/database"
	"github.com/forbole/juno/v5/types"
)

// HandleMsg represents a message handler that stores the given message inside the proper database table
func HandleMsg(
	index int, msg sdk.Msg, tx *types.Tx,
	parseAddresses MessageAddressesParser, cdc codec.Codec, db database.Database,
) error {

	// Get the involved addresses
	addresses, err := parseAddresses(tx)
	if err != nil {
		return err
	}

	// Marshal the value properly
	bz, err := cdc.MarshalJSON(msg)
	if err != nil {
		return err
	}

	msgLabel := GetMsgFromTypeURL(proto.MessageType(proto.MessageName(msg)).String())

	// Save message type
	err = db.SaveMessageType(types.NewMessageType(index,
		proto.MessageName(msg),
		GetModuleNameFromTypeURL(proto.MessageName(msg)),
		msgLabel,
		tx.Height))

	if err != nil {
		return err
	}

	// Handle MsgRecvPacket data object
	if msgIBC, ok := msg.(*channeltypes.MsgRecvPacket); ok {
		trimMessageString := TrimLastChar(string(bz))
		trimDataString := string(msgIBC.Packet.Data)[1:]
		return db.SaveMessage(types.NewMessage(
			tx.TxHash,
			index,
			proto.MessageName(msg),
			fmt.Sprintf("%s,%s", trimMessageString, trimDataString),
			addresses,
			tx.Height,
		))
	}

	return db.SaveMessage(types.NewMessage(
		tx.TxHash,
		index,
		proto.MessageName(msg),
		string(bz),
		addresses,
		tx.Height,
	))
}

func GetModuleNameFromTypeURL(input string) string {
	moduleName := strings.Split(input, ".")
	if len(moduleName) > 1 {
		if strings.Contains(moduleName[0], "cosmos") {
			return moduleName[1] // e.g. "cosmos.bank.v1beta1.MsgSend" => "bank"
		} else if strings.Contains(moduleName[0], "ibc") {
			return fmt.Sprintf("%s %s %s", moduleName[0], moduleName[1], moduleName[2]) // e.g. "ibc.core.client.v1.MsgUpdateClient" => "ibc core client"
		} else {
			return fmt.Sprintf("%s %s", moduleName[0], moduleName[1]) // e.g. "cosmwasm.wasm.v1.MsgExecuteContract" => "cosmwasm wasm"
		}

	}

	return ""
}

func GetMsgFromTypeURL(input string) string {
	messageName := strings.Split(input, ".")
	if len(messageName) > 1 {
		return messageName[1] // e.g. "*types.MsgSend" => "MsgSend"
	}
	return ""
}
