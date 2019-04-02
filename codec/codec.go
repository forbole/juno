package codec

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

// Codec is the application-wide Amino codec and is initialized upon package
// loading.
var Codec *codec.Codec

func init() {
	Codec = codec.New()

	auth.RegisterCodec(Codec)
	sdk.RegisterCodec(Codec)
	codec.RegisterCrypto(Codec)
}
