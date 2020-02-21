package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SdkConfigSetup represents a method that allows to customize the given sdk.Config.
// This should be used to set custom Bech32 addresses prefixes and other app-related configurations.
type SdkConfigSetup func(*sdk.Config)

// Handy implementation of SdkConfigSetup that performs no operations
func EmptySetup(*sdk.Config) {}

// CodecBuilder represents a function that is used to return the proper application codec.
type CodecBuilder func() *codec.Codec
