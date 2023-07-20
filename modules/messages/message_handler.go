package messages

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"

	// clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	// "github.com/cosmos/ibc-go/v7/modules/core/exported"
	"github.com/forbole/juno/v5/database"
	"github.com/forbole/juno/v5/types"
	"github.com/forbole/juno/v5/types/config"
)

// HandleMsg represents a message handler that stores the given message inside the proper database table
func HandleMsg(
	index int, msg sdk.Msg, tx *types.Tx,
	parseAddresses MessageAddressesParser, cdc codec.Codec, db database.Database,
) error {

	// Get the involved addresses
	addresses, err := parseAddresses(tx, config.Cfg.Chain.Bech32Prefix)
	if err != nil {
		return err
	}

	// Marshal the value properly
	bz, err := cdc.MarshalJSON(msg)
	if err != nil {
		return err
	}

	if msgIBC, ok := msg.(*transfertypes.MsgTransfer); ok {
		return db.SaveIBCMsgTransferRelationship(types.NewIBCMsgTransferRelationship(tx.TxHash, index, msgIBC.SourcePort, msgIBC.SourceChannel,
			msgIBC.Sender, msgIBC.Receiver, tx.Height))
	}

	// Handle MsgRecvPacket data object
	if msgIBC, ok := msg.(*channeltypes.MsgRecvPacket); ok {
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

		var data transfertypes.FungibleTokenPacketData
		if err := transfertypes.ModuleCdc.UnmarshalJSON(msgIBC.Packet.Data, &data); err != nil {
			// The packet data is not a FungibleTokenPacketData, so nothing to update
			return nil
		}

		return db.SaveIBCMsgRecvPacketRelationship(types.NewIBCMsgRecvPacketRelationship(tx.TxHash, index, string(msgIBC.Packet.Data), fmt.Sprint(msgIBC.Packet.Sequence), msgIBC.Packet.SourcePort, msgIBC.Packet.SourceChannel,
			msgIBC.Packet.DestinationPort, msgIBC.Packet.DestinationChannel, data.Sender, data.Receiver, tx.Height))
	}

	if msgIBC, ok := msg.(*channeltypes.MsgAcknowledgement); ok {
		var data transfertypes.FungibleTokenPacketData
		if err := transfertypes.ModuleCdc.UnmarshalJSON(msgIBC.Packet.Data, &data); err != nil {
			// The packet data is not a FungibleTokenPacketData, so nothing to update
			return nil
		}

		return db.SaveIBCMsgAcknowledgementRelationship(types.NewIBCMsgAcknowledgementRelationship(tx.TxHash, index, string(msgIBC.Packet.Data), fmt.Sprint(msgIBC.Packet.Sequence), msgIBC.Packet.SourcePort, msgIBC.Packet.SourceChannel,
			msgIBC.Packet.DestinationPort, msgIBC.Packet.DestinationChannel, data.Sender, data.Receiver, tx.Height))
	}

	// if msgIBC, ok := msg.(*clienttypes.MsgCreateClient); ok {
	// 	var clientState exported.ClientState
	// 	if err := cdc.UnmarshalJSON([]byte(msgIBC.ClientState.Value), clientState); err != nil {
	// 		return err
	// 	}
	// 	return db.SaveIBCClientMessageRelationship(types.NewIBCClientMessageRelationship(tx.TxHash, clientState., tx.Height))
	// }

	// if msgIBC, ok := msg.(*clienttypes.MsgUpdateClient); ok {
	// 	return db.SaveIBCClientMessageRelationship(types.NewIBCClientMessageRelationship(tx.TxHash, msgIBC.Signer, tx.Height))
	// }

	// if msgIBC, ok := msg.(*clienttypes.MsgUpgradeClient); ok {
	// 	return db.SaveIBCClientMessageRelationship(types.NewIBCClientMessageRelationship(tx.TxHash, msgIBC.Signer, tx.Height))
	// }

	return db.SaveMessage(types.NewMessage(
		tx.TxHash,
		index,
		proto.MessageName(msg),
		string(bz),
		addresses,
		tx.Height,
	))
}
