package client

import (
	"google.golang.org/grpc"

	"github.com/desmos-labs/juno/types"
)

// CreateGrpcConnection creates a new gRPC client connection from the given configuration
func CreateGrpcConnection(cfg types.Config) (*grpc.ClientConn, error) {
	gprConfig := cfg.GetGrpcConfig()

	var grpcOpts []grpc.DialOption
	if gprConfig.Insecure {
		grpcOpts = append(grpcOpts, grpc.WithInsecure())
	}

	return grpc.Dial(gprConfig.Address, grpcOpts...)
}
