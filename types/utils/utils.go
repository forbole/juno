package utils

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	abci "github.com/tendermint/tendermint/abci/types"
	tmcrypto "github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/types"
)

// ConvertValidatorAddressToBech32String converts the given validator address to its Bech32 string representation
func ConvertValidatorAddressToBech32String(address types.Address) string {
	return sdk.ConsAddress(address).String()
}

// ConvertValidatorPubKeyToBech32String converts the given pubKey to a Bech32 string
func ConvertValidatorPubKeyToBech32String(pubKey tmcrypto.PubKey) (string, error) {
	bech32Prefix := sdk.GetConfig().GetBech32ConsensusPubPrefix()
	return bech32.ConvertAndEncode(bech32Prefix, pubKey.Bytes())
}

func FindEventByType(events []abci.Event, eventType string) (abci.Event, error) {
	for _, event := range events {
		if event.Type == eventType {
			return event, nil
		}
	}

	return abci.Event{}, fmt.Errorf("no event with type %s found", eventType)
}

func FindEventsByType(events []abci.Event, eventType string) []abci.Event {
	var found []abci.Event
	for _, event := range events {
		if event.Type == eventType {
			found = append(found, event)
		}
	}

	return found
}

func FindAttributeByKey(event abci.Event, attrKey string) (abci.EventAttribute, error) {
	for _, attr := range event.Attributes {
		if string(attr.Key) == attrKey {
			return attr, nil
		}
	}

	return abci.EventAttribute{}, fmt.Errorf("no attribute with key %s found inside event with type %s", attrKey, event.Type)
}

func MaxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
