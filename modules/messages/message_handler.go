package messages

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	"github.com/gogo/protobuf/proto"

	"github.com/forbole/juno/v4/database"
	"github.com/forbole/juno/v4/types"
)

// HandleMsg represents a message handler that stores the given message inside the proper database table
func HandleMsg(
	index int, msg sdk.Msg, tx *types.Tx,
	parseAddresses MessageAddressesParser, cdc codec.Codec, db database.Database,
) error {

	// Get the involved addresses
	addresses, err := parseAddresses(cdc, msg)
	if err != nil {
		return err
	}

	// Marshal the value properly
	bz, err := cdc.MarshalJSON(msg)
	if err != nil {
		return err
	}

	if msgIBC, ok := msg.(*channeltypes.MsgRecvPacket); ok {
		// packet := types.Packet{
		// 	Sequence:           msgIBC.Packet.Sequence,
		// 	SourcePort:         msgIBC.Packet.SourcePort,
		// 	SourceChannel:      msgIBC.Packet.SourceChannel,
		// 	DestinationPort:    msgIBC.Packet.DestinationPort,
		// 	DestinationChannel: msgIBC.Packet.DestinationChannel,
		// 	Data:               string(msgIBC.Packet.Data),
		// 	TimeoutHeight:      msgIBC.Packet.TimeoutHeight,
		// 	TimeoutTimestamp:   msgIBC.Packet.TimeoutTimestamp,
		// }
		// msgRecvPacket = types.MsgRecvPacket{
		// 	Packet:          packet,
		// 	ProofCommitment: msgIBC.ProofCommitment,
		// 	ProofHeight:     msgIBC.ProofHeight,
		// 	Signer:          msgIBC.Signer,
		// }

		// msgString := fmt.Sprintf("%#v", msgRecvPacket)
		// fmt.Printf("\n\n msgString %s \n\n", msgString)
		// return db.SaveMessage(types.NewMessage(
		// 	tx.TxHash,
		// 	index,
		// 	proto.MessageName(msg),
		// 	msgString,
		// 	addresses,
		// 	tx.Height,
		// ))
		datas := string(msgIBC.Packet.Data)
		messageString := fmt.Sprintf("%s,%s", string(bz), datas)
		fmt.Printf("\n\n messageString %s \n\n", messageString)
		return db.SaveMessage(types.NewMessage(
			tx.TxHash,
			index,
			proto.MessageName(msg),
			messageString,
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
