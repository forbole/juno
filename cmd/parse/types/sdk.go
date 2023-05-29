package types

import (
	"github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/forbole/juno/v5/types/config"
)

// SdkConfigSetup represents a method that allows to customize the given sdk.Config.
// This should be used to set custom Bech32 addresses prefixes and other app-related configurations.
type SdkConfigSetup func(config config.Config, sdkConfig *sdk.Config)

// DefaultConfigSetup represents a handy implementation of SdkConfigSetup that simply setups the prefix
// inside the configuration
func DefaultConfigSetup(cfg config.Config, sdkConfig *sdk.Config) {
	prefixes := cfg.Chain.Bech32Prefix

	bech32PrefixesAccPub := make([]string, 0, len(prefixes))
	bech32PrefixesValAddr := make([]string, 0, len(prefixes))
	bech32PrefixesValPub := make([]string, 0, len(prefixes))
	bech32PrefixesConsAddr := make([]string, 0, len(prefixes))
	bech32PrefixesConsPub := make([]string, 0, len(prefixes))

	for _, prefix := range prefixes {
		bech32PrefixesAccPub = append(bech32PrefixesAccPub, prefix+"pub")
		bech32PrefixesValAddr = append(bech32PrefixesValAddr, prefix+"valoper")
		bech32PrefixesValPub = append(bech32PrefixesValPub, prefix+"valoperpub")
		bech32PrefixesConsAddr = append(bech32PrefixesConsAddr, prefix+"valcons")
		bech32PrefixesConsPub = append(bech32PrefixesConsPub, prefix+"valconspub")
	}

	sdkConfig.SetBech32PrefixesForAccount(
		prefixes,
		bech32PrefixesAccPub,
	)
	sdkConfig.SetBech32PrefixesForValidator(
		bech32PrefixesValAddr,
		bech32PrefixesValPub,
	)
	sdkConfig.SetBech32PrefixesForConsensusNode(
		bech32PrefixesConsAddr,
		bech32PrefixesConsPub,
	)
}

// -----------------------------------------------------------------

// EncodingConfigBuilder represents a function that is used to return the proper encoding config.
type EncodingConfigBuilder func() params.EncodingConfig
