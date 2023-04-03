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
	var msgRecvPacket types.MsgRecvPacket
	if msgIBC, ok := msg.(*channeltypes.MsgRecvPacket); ok {
		packet := types.Packet{
			Sequence:           msgIBC.Packet.Sequence,
			SourcePort:         msgIBC.Packet.SourcePort,
			SourceChannel:      msgIBC.Packet.SourceChannel,
			DestinationPort:    msgIBC.Packet.DestinationPort,
			DestinationChannel: msgIBC.Packet.DestinationChannel,
			Data:               string(msgIBC.Packet.Data),
			TimeoutHeight:      msgIBC.Packet.TimeoutHeight,
			TimeoutTimestamp:   msgIBC.Packet.TimeoutTimestamp,
		}
		msgRecvPacket = types.MsgRecvPacket{
			Packet:          packet,
			ProofCommitment: msgIBC.ProofCommitment,
			ProofHeight:     msgIBC.ProofHeight,
			Signer:          msgIBC.Signer,
		}

		return db.SaveMessage(types.NewMessage(
			tx.TxHash,
			index,
			proto.MessageName(msg),
			fmt.Sprintf("%#v", msgRecvPacket),
			addresses,
			tx.Height,
		))
	}

	// Marshal the value properly
	bz, err := cdc.MarshalJSON(msg)
	if err != nil {
		return err
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
