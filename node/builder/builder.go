package builder

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/simapp/params"

	"github.com/saifullah619/juno/v3/node"
	nodeconfig "github.com/saifullah619/juno/v3/node/config"
	"github.com/saifullah619/juno/v3/node/local"
	"github.com/saifullah619/juno/v3/node/remote"
)

func BuildNode(cfg nodeconfig.Config, encodingConfig *params.EncodingConfig) (node.Node, error) {
	switch cfg.Type {
	case nodeconfig.TypeRemote:
		return remote.NewNode(cfg.Details.(*remote.Details), encodingConfig.Codec)
	case nodeconfig.TypeLocal:
		return local.NewNode(cfg.Details.(*local.Details), encodingConfig.TxConfig, encodingConfig.Codec)
	case nodeconfig.TypeNone:
		return nil, nil

	default:
		return nil, fmt.Errorf("invalid node type: %s", cfg.Type)
	}
}
