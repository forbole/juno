package types

import (
	"github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/forbole/juno/v4/types/config"
)

// SdkConfigSetup represents a method that allows to customize the given sdk.Config.
// This should be used to set custom Bech32 addresses prefixes and other app-related configurations.
type SdkConfigSetup func(config config.Config, sdkConfig *sdk.Config)

// DefaultConfigSetup represents a handy implementation of SdkConfigSetup that simply setups the prefix
// inside the configuration
func DefaultConfigSetup(cfg config.Config, sdkConfig *sdk.Config) {
	prefix := cfg.Chain.Bech32Prefix
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
