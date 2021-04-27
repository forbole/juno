package types

import (
	"github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SdkConfigSetup represents a method that allows to customize the given sdk.Config.
// This should be used to set custom Bech32 addresses prefixes and other app-related configurations.
type SdkConfigSetup func(config *Config, sdkConfig *sdk.Config)

// DefaultSetup represents a handy implementation of SdkConfigSetup that simply setups the prefix
// inside the configuration
func DefaultSetup(cfg *Config, sdkConfig *sdk.Config) {
	prefix := cfg.Cosmos.Prefix
	sdkConfig.SetBech32PrefixForAccount(
		prefix,
		prefix+sdk.PrefixPublic,
	)
	sdkConfig.SetBech32PrefixForValidator(
		prefix+sdk.PrefixValidator+sdk.PrefixOperator,
		prefix+sdk.PrefixValidator+sdk.PrefixOperator+sdk.PrefixPublic,
	)
	sdkConfig.SetBech32PrefixForConsensusNode(
		prefix+sdk.PrefixValidator+sdk.PrefixConsensus,
		prefix+sdk.PrefixValidator+sdk.PrefixConsensus+sdk.PrefixPublic,
	)
}

// -----------------------------------------------------------------

// EncodingConfigBuilder represents a function that is used to return the proper encoding config.
type EncodingConfigBuilder func() params.EncodingConfig
