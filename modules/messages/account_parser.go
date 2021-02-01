package messages

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	ibctransfertypes "github.com/cosmos/cosmos-sdk/x/ibc/applications/transfer/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// MessageAddressesParser represents a function that extracts all the
// involved addresses from a provided message (both accounts and validators)
type MessageAddressesParser = func(cdc codec.Marshaler, msg sdk.Msg) ([]string, error)

// CosmosMessageAddressesParser represents a MessageAddressesParser that parses a
// Cosmos message and returns all the involved addresses (both accounts and validators)
// nolint: gocyclo
func CosmosMessageAddressesParser(cdc codec.Marshaler, cosmosMsg sdk.Msg) ([]string, error) {
	switch msg := cosmosMsg.(type) {
	case *banktypes.MsgSend:
		return []string{msg.ToAddress, msg.FromAddress}, nil

	case *banktypes.MsgMultiSend:
		var addresses []string
		for _, i := range msg.Inputs {
			addresses = append(addresses, i.Address)
		}
		for _, o := range msg.Outputs {
			addresses = append(addresses, o.Address)
		}
		return addresses, nil

	case *crisistypes.MsgVerifyInvariant:
		return []string{msg.Sender}, nil

	case *distrtypes.MsgSetWithdrawAddress:
		return []string{msg.DelegatorAddress, msg.WithdrawAddress}, nil

	case *distrtypes.MsgWithdrawDelegatorReward:
		return []string{msg.DelegatorAddress, msg.ValidatorAddress}, nil

	case *distrtypes.MsgWithdrawValidatorCommission:
		return []string{msg.ValidatorAddress}, nil

	case *distrtypes.MsgFundCommunityPool:
		return []string{msg.Depositor}, nil

	case *evidencetypes.MsgSubmitEvidence:
		return []string{msg.Submitter}, nil

	case *govtypes.MsgSubmitProposal:
		addresses := []string{msg.Proposer}

		var content govtypes.Content
		err := cdc.UnpackAny(msg.Content, &content)
		if err != nil {
			return nil, err
		}

		// Get addresses from contents
		switch content := content.(type) {
		case *distrtypes.CommunityPoolSpendProposal:
			addresses = append(addresses, content.Recipient)
		}

		return addresses, nil

	case *govtypes.MsgDeposit:
		return []string{msg.Depositor}, nil

	case *govtypes.MsgVote:
		return []string{msg.Voter}, nil

	case *ibctransfertypes.MsgTransfer:
		return []string{msg.Sender, msg.Receiver}, nil

	case *slashingtypes.MsgUnjail:
		return []string{msg.ValidatorAddr}, nil

	case *stakingtypes.MsgCreateValidator:
		return []string{msg.ValidatorAddress, msg.DelegatorAddress}, nil

	case *stakingtypes.MsgEditValidator:
		return []string{msg.ValidatorAddress}, nil

	case *stakingtypes.MsgDelegate:
		return []string{msg.DelegatorAddress, msg.ValidatorAddress}, nil

	case *stakingtypes.MsgBeginRedelegate:
		return []string{msg.DelegatorAddress, msg.ValidatorSrcAddress, msg.ValidatorDstAddress}, nil

	case *stakingtypes.MsgUndelegate:
		return []string{msg.DelegatorAddress, msg.ValidatorAddress}, nil
	}

	return []string{}, fmt.Errorf("message type not supported: %s", cosmosMsg.Type())
}
