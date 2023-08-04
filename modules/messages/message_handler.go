package messages

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/gogo/protobuf/proto"

	transfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v4/modules/core/04-channel/types"

	"github.com/forbole/juno/v4/database"
	"github.com/forbole/juno/v4/types"
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

	// Handle IBC MsgTransfer data from events
	if msgIBC, ok := msg.(*transfertypes.MsgTransfer); ok {
		packetData, packetSequence, destinationPort, destinationChannel, err := parseTxEvents(tx)
		if err != nil {
			fmt.Printf("error while parsing events for MsgTransfer ibc relationship, tx: %s, error: %s ", tx.TxHash, err)
		} else {
			// Save IBC message relationship inside message_ibc_relationship table
			db.SaveIBCMsgRelationship(types.NewIBCMsgRelationship(tx.TxHash, index, proto.MessageName(msg),
				packetData, packetSequence, msgIBC.SourcePort, msgIBC.SourceChannel, destinationPort,
				destinationChannel, msgIBC.Sender, msgIBC.Receiver, tx.Height))
		}

	}

	// Handle IBC MsgRecvPacket data object
	if msgIBC, ok := msg.(*channeltypes.MsgRecvPacket); ok {
		// Parse MsgRecvPacket Data and store in message table
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

		// Parse sender and receiver address for IBC relationship
		sender, receiver, err := parsePacketData(msgIBC.Packet.Data, tx)
		if err != nil {
			return fmt.Errorf("error while unmarshalling sender and receiver address for MsgRecvPacket ibc relationship, tx: %s, error: %s ", tx.TxHash, err)
		}

		// Save IBC message relationship inside message_ibc_relationship table
		return db.SaveIBCMsgRelationship(types.NewIBCMsgRelationship(tx.TxHash, index, proto.MessageName(msg),
			string(msgIBC.Packet.Data), fmt.Sprint(msgIBC.Packet.Sequence), msgIBC.Packet.SourcePort, msgIBC.Packet.SourceChannel,
			msgIBC.Packet.DestinationPort, msgIBC.Packet.DestinationChannel, sender, receiver, tx.Height))
	}

	// Handle IBC MsgAcknowledgement data object
	if msgIBC, ok := msg.(*channeltypes.MsgAcknowledgement); ok {
		// Parse sender and receiver address for IBC relationship
		sender, receiver, err := parsePacketData(msgIBC.Packet.Data, tx)
		if err != nil {
			fmt.Printf("error while unmarshalling sender and receiver address for MsgAcknowledgement ibc relationship, tx: %s, error: %s ", tx.TxHash, err)
		} else {
			// Save IBC message relationship inside message_ibc_relationship table
			db.SaveIBCMsgRelationship(types.NewIBCMsgRelationship(tx.TxHash, index, proto.MessageName(msg),
				string(msgIBC.Packet.Data), fmt.Sprint(msgIBC.Packet.Sequence), msgIBC.Packet.SourcePort, msgIBC.Packet.SourceChannel,
				msgIBC.Packet.DestinationPort, msgIBC.Packet.DestinationChannel, sender, receiver, tx.Height))
		}
	}

	// Save new message inside message table
	return db.SaveMessage(types.NewMessage(
		tx.TxHash,
		index,
		proto.MessageName(msg),
		string(bz),
		addresses,
		tx.Height,
	))
}

func parseTxEvents(tx *types.Tx) (string, string, string, string, error) {
	var packetData, packetSequence, destinationPort, destinationChannel string

	for _, event := range tx.Events {
		if event.Type == channeltypes.EventTypeSendPacket {
			for _, attribute := range event.Attributes {
				if string(attribute.Key) == channeltypes.AttributeKeyData {
					packetData = string(attribute.Value)
				}
				if string(attribute.Key) == channeltypes.AttributeKeySequence {
					packetSequence = string(attribute.Value)
				}
				if string(attribute.Key) == channeltypes.AttributeKeyDstPort {
					destinationPort = string(attribute.Value)
				}
				if string(attribute.Key) == channeltypes.AttributeKeyDstChannel {
					destinationChannel = string(attribute.Value)
				}
			}
		}
	}

	if len(packetData) > 0 && len(packetSequence) > 0 && len(destinationChannel) > 0 && len(destinationPort) > 0 {
		return packetData, packetSequence, destinationPort, destinationChannel, nil
	} else {
		return "", "", "", "", fmt.Errorf("couldn't parse ibc message relationship details from events for tx at height: %d", tx.Height)
	}
}

func parsePacketData(packetData []byte, tx *types.Tx) (string, string, error) {
	// Parse sender and receiver address for ibc relationship
	var data transfertypes.FungibleTokenPacketData
	if err := transfertypes.ModuleCdc.UnmarshalJSON(packetData, &data); err != nil {
		// If packetData is not a FungibleTokenPacketData type, parse sender
		// and receiver addresses from events
		var sender, receiver sdk.AccAddress
		for _, event := range tx.Events {
			if event.Type == transfertypes.EventTypeTransfer {
				for _, attribute := range event.Attributes {
					if string(attribute.Key) == banktypes.AttributeKeySender {
						// check if event value is sdk address
						sender, err = sdk.AccAddressFromBech32(string(attribute.Value))
						if err != nil {
							// skip if value is not sdk address
							continue
						}
					} else if string(attribute.Key) == banktypes.AttributeKeyReceiver {
						// check if event value is sdk address
						receiver, err = sdk.AccAddressFromBech32(string(attribute.Value))
						if err != nil {
							// skip if value is not sdk address
							continue
						}
					}
				}
			} else if event.Type == sdk.EventTypeMessage {
				for _, attribute := range event.Attributes {
					if string(attribute.Key) == banktypes.AttributeKeySender {
						// check if event value is sdk address
						sender, err = sdk.AccAddressFromBech32(string(attribute.Value))
						if err != nil {
							// skip if value is not sdk address
							continue
						}
					} else if string(attribute.Key) == banktypes.AttributeKeyReceiver {
						// check if event value is sdk address
						receiver, err = sdk.AccAddressFromBech32(string(attribute.Value))
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
