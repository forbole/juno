package utils

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/types"
)

// ConvertValidatorAddressToString converts the given validator address to its string representation
func ConvertValidatorAddressToString(address types.Address) string {
	return sdk.ConsAddress(address).String()
}
