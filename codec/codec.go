package codec

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
)

// Codec is the application-wide Amino codec and is initialized upon package
// loading.
var Codec *codec.Codec

func init() {
	Codec = simapp.MakeCodec()
	Codec.Seal()
}
