package builder

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/simapp/params"

	"github.com/forbole/juno/v2/node"
	nodeconfig "github.com/forbole/juno/v2/node/config"
	"github.com/forbole/juno/v2/node/local"
	"github.com/forbole/juno/v2/node/remote"
)

func BuildNode(cfg nodeconfig.Config, encodingConfig *params.EncodingConfig) (node.Node, error) {
	switch cfg.Type {
	case nodeconfig.TypeRemote:
		return remote.NewNode(cfg.Details.(*remote.Details), encodingConfig.Marshaler)
	case nodeconfig.TypeLocal:
		return local.NewNode(cfg.Details.(*local.Details), encodingConfig.TxConfig, encodingConfig.Marshaler)

	default:
		return nil, fmt.Errorf("invalid node type: %s", cfg.Type)
	}
}
