package remote

import (
	"context"

	"google.golang.org/grpc"

	"github.com/desmos-labs/juno/node"
)

var (
	_ node.Keeper = &Keeper{}
)

// Keeper implements the keeper.Keeper interface relying on a GRPC connection
type Keeper struct {
	Ctx      context.Context
	GrpcConn *grpc.ClientConn
}

// NewKeeper returns a new Keeper instance
func NewKeeper(config *GRPCConfig) (*Keeper, error) {
	return &Keeper{
		Ctx:      context.Background(),
		GrpcConn: MustCreateGrpcConnection(config),
	}, nil
}

// Type implements keeper.Type
func (k Keeper) Type() string {
	return node.RemoteKeeper
}
