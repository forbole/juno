package client

import (
	"github.com/desmos-labs/juno/types"
	"google.golang.org/grpc"
)

// CreateGrpcConnection creates a new gRPC client connection from the given configuration
func CreateGrpcConnection(cfg *types.Config) (*grpc.ClientConn, error) {
	var grpcOpts []grpc.DialOption
	if cfg.Grpc.Insecure {
		grpcOpts = append(grpcOpts, grpc.WithInsecure())
	}

	return grpc.Dial(cfg.Grpc.Address, grpcOpts...)
}
