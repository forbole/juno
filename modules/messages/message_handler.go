package messages

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
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

	// Handle ibc MsgTransfer
	if msgIBC, ok := msg.(*transfertypes.MsgTransfer); ok {
		var packetData, packetSequence, destinationPort, destinationChannel string

		for _, event := range tx.Events {
			if event.Type == channeltypes.EventTypeSendPacket {
				for _, attribute := range event.Attributes {
					if attribute.Key == channeltypes.AttributeKeyData {
						packetData = attribute.Value
					}
					if attribute.Key == channeltypes.AttributeKeySequence {
						packetSequence = attribute.Value
					}
					if attribute.Key == channeltypes.AttributeKeyDstPort {
						destinationPort = attribute.Value
					}
					if attribute.Key == channeltypes.AttributeKeyDstChannel {
						destinationChannel = attribute.Value
					}
				}

			}
		}
		db.SaveIBCMsgRelationship(types.NewIBCMsgRelationship(tx.TxHash, index, proto.MessageName(msg),
			packetData, packetSequence, msgIBC.SourcePort, msgIBC.SourceChannel, destinationPort,
			destinationChannel, msgIBC.Sender, msgIBC.Receiver, tx.Height))
	}

	// Handle ibc MsgRecvPacket data object
	if msgIBC, ok := msg.(*channeltypes.MsgRecvPacket); ok {
		// parse MsgRecvPacket Data and store in message table
		trimMessageString := TrimLastChar(string(bz))
		trimDataString := string(msgIBC.Packet.Data)[1:]
		err := db.SaveMessage(types.NewMessage(
			tx.TxHash,
			index,
			proto.MessageName(msg),
			fmt.Sprintf("%s,%s", trimMessageString, trimDataString),
			addresses,
			tx.Height,
		))
		if err != nil {
			return err
		}

		// parse sender and receiver address for ibc relationship
		sender, receiver, err := parsePacketData(msgIBC.Packet.Data, tx)
		if err != nil {
			fmt.Printf("error while unmarshalling sender and receiver address for MsgRecvPacket ibc relationship, tx: %s, error: %s ", tx.TxHash, err)
		}

		return db.SaveIBCMsgRelationship(types.NewIBCMsgRelationship(tx.TxHash, index, proto.MessageName(msg),
			string(msgIBC.Packet.Data), fmt.Sprint(msgIBC.Packet.Sequence), msgIBC.Packet.SourcePort, msgIBC.Packet.SourceChannel,
			msgIBC.Packet.DestinationPort, msgIBC.Packet.DestinationChannel, sender, receiver, tx.Height))
	}

	// Handle ibc MsgAcknowledgement data object
	if msgIBC, ok := msg.(*channeltypes.MsgAcknowledgement); ok {
		// parse sender and receiver address for ibc relationship
		sender, receiver, err := parsePacketData(msgIBC.Packet.Data, tx)
		if err != nil {
			fmt.Printf("error while unmarshalling sender and receiver address for MsgAcknowledgement ibc relationship, tx: %s, error: %s ", tx.TxHash, err)
		}

		db.SaveIBCMsgRelationship(types.NewIBCMsgRelationship(tx.TxHash, index, proto.MessageName(msg),
			string(msgIBC.Packet.Data), fmt.Sprint(msgIBC.Packet.Sequence), msgIBC.Packet.SourcePort, msgIBC.Packet.SourceChannel,
			msgIBC.Packet.DestinationPort, msgIBC.Packet.DestinationChannel, sender, receiver, tx.Height))
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

func parsePacketData(packetData []byte, tx *types.Tx) (string, string, error) {
	// parse sender and receiver address for ibc relationship
	var data transfertypes.FungibleTokenPacketData
	if err := transfertypes.ModuleCdc.UnmarshalJSON(packetData, &data); err != nil {
		var sender, receiver sdk.AccAddress
		for _, event := range tx.Events {
			if event.Type == "message" {
				for _, attribute := range event.Attributes {
					if attribute.Key == "sender" {
						// check if event value is sdk address
						sender, err = sdk.AccAddressFromBech32(attribute.Value)
						if err != nil {
							// skip if value is not sdk address
							continue
						}
					} else if attribute.Key == "receiver" {
						// check if event value is sdk address
						receiver, err = sdk.AccAddressFromBech32(attribute.Value)
						if err != nil {
							// skip if value is not sdk address
							continue
						}
					}

				}
			}

		}
		return sender.String(), receiver.String(), err
	}
	return data.Sender, data.Receiver, nil
}
