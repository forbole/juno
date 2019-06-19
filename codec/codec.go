package codec

import (
	"github.com/cosmos/cosmos-sdk/cmd/gaia/app"
	"github.com/cosmos/cosmos-sdk/codec"
)

// Codec is the application-wide Amino codec and is initialized upon package
// loading.
var Codec *codec.Codec

func init() {
	Codec = app.MakeCodec()
	Codec.Seal()
}
