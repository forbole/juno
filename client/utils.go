package client

import (
	"strconv"

	"github.com/desmos-labs/juno/types"

	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// GetHeightRequestHeader returns the grpc.CallOption to query the state at a given height
func GetHeightRequestHeader(height int64) grpc.CallOption {
	header := metadata.New(map[string]string{
		grpctypes.GRPCBlockHeightHeader: strconv.FormatInt(height, 10),
	})
	return grpc.Header(&header)
}

// MustCreateGrpcConnection creates a new gRPC connection using the provided configuration and panics on error
func MustCreateGrpcConnection(cfg types.Config) *grpc.ClientConn {
	grpConnection, err := CreateGrpcConnection(cfg)
	if err != nil {
		panic(err)
	}
	return grpConnection
}

// CreateGrpcConnection creates a new gRPC client connection from the given configuration
func CreateGrpcConnection(cfg types.Config) (*grpc.ClientConn, error) {
	gprConfig := cfg.GetGrpcConfig()

	var grpcOpts []grpc.DialOption
	if gprConfig.IsInsecure() {
		grpcOpts = append(grpcOpts, grpc.WithInsecure())
	}

	return grpc.Dial(gprConfig.GetAddress(), grpcOpts...)
}
