package messages

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	"github.com/gogo/protobuf/proto"

	didtypes "github.com/cheqd/cheqd-node/x/did/types"
	resources "github.com/cheqd/cheqd-node/x/resource/types"
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

	// Handle involved addresses for Cheqd' did and resource module
	var involvedAddresses []string
	switch msg.(type) {
	case *didtypes.MsgCreateDidDoc:
		involvedAddresses = append(involvedAddresses, tx.FeePayer().String(), tx.FeeGranter().String())
		return db.SaveMessage(types.NewMessage(
			tx.TxHash,
			index,
			proto.MessageName(msg),
			string(bz),
			involvedAddresses,
			tx.Height,
		))
	case *didtypes.MsgUpdateDidDoc:
		involvedAddresses = append(involvedAddresses, tx.FeePayer().String(), tx.FeeGranter().String())
		return db.SaveMessage(types.NewMessage(
			tx.TxHash,
			index,
			proto.MessageName(msg),
			string(bz),
			involvedAddresses,
			tx.Height,
		))
	case *didtypes.MsgDeactivateDidDoc:
		involvedAddresses = append(involvedAddresses, tx.FeePayer().String(), tx.FeeGranter().String())
		return db.SaveMessage(types.NewMessage(
			tx.TxHash,
			index,
			proto.MessageName(msg),
			string(bz),
			involvedAddresses,
			tx.Height,
		))
	case *resources.MsgCreateResource:
		involvedAddresses = append(involvedAddresses, tx.FeePayer().String(), tx.FeeGranter().String())
		return db.SaveMessage(types.NewMessage(
			tx.TxHash,
			index,
			proto.MessageName(msg),
			string(bz),
			involvedAddresses,
			tx.Height,
		))
	// Handle MsgRecvPacket data object
	case *channeltypes.MsgRecvPacket:
		trimMessageString := TrimLastChar(string(bz))
		trimDataString := string(msg.(*channeltypes.MsgRecvPacket).Packet.Data)[1:]
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
