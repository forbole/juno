package client

import (
	"github.com/desmos-labs/juno/config"
	"google.golang.org/grpc"
)

// CreateGrpcConnection creates a new gRPC client connection from the given configuration
func CreateGrpcConnection(cfg *config.Config) (*grpc.ClientConn, error) {
	var grpcOpts []grpc.DialOption
	if cfg.GrpcConfig.Insecure {
		grpcOpts = append(grpcOpts, grpc.WithInsecure())
	}

	return grpc.Dial(cfg.GrpcConfig.Address, grpcOpts...)
}
