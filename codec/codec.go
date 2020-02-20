package codec

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	postTypes "github.com/desmos-labs/desmos/x/posts"
)

// Codec is the application-wide Amino codec and is initialized upon package
// loading.
var Codec *codec.Codec
var _ sdk.Msg = postTypes.MsgCreatePost{}
var _ sdk.Msg = postTypes.MsgEditPost{}

func init() {
	Codec = simapp.MakeCodec()
	Codec.RegisterConcrete(postTypes.MsgCreatePost{}, "desmos/MsgCreatePost", nil)
	Codec.RegisterConcrete(postTypes.MsgEditPost{}, "desmos/MsgEditPost", nil)
	Codec.Seal()
}
