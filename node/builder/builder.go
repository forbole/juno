package builder

import (
	"fmt"

	"cosmossdk.io/simapp/params"

	"github.com/emrahm/juno/v5/node"
	nodeconfig "github.com/emrahm/juno/v5/node/config"
	"github.com/emrahm/juno/v5/node/remote"
)

func BuildNode(cfg nodeconfig.Config, encodingConfig *params.EncodingConfig) (node.Node, error) {
	switch cfg.Type {
	case nodeconfig.TypeRemote:
		return remote.NewNode(cfg.Details.(*remote.Details), encodingConfig.Codec)
	case nodeconfig.TypeNone:
		return nil, nil

	default:
		return nil, fmt.Errorf("invalid node type: %s", cfg.Type)
	}
}
