package sdk

import (
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/cosmos/cosmos-sdk/codec/types"
	auth "github.com/cosmos/cosmos-sdk/x/auth/types"

	"google.golang.org/grpc"
)

type Client struct {
	authClient   auth.QueryClient
	bankClient   banktypes.QueryClient
	encodeConfig types.InterfaceRegistry

	endpoint string
}

// NewClient returns the client to call Cosmos RPC.
func NewClient(endpoint string) (*Client, error) {
	// instantiate gRPC connection
	grpcConn, err := grpc.Dial(endpoint, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	// create interface registry, and register used modules types
	interfaceRegistry := types.NewInterfaceRegistry()
	auth.RegisterInterfaces(interfaceRegistry)
	banktypes.RegisterInterfaces(interfaceRegistry)
	cryptocodec.RegisterInterfaces(interfaceRegistry)
	return &Client{
		authClient:   auth.NewQueryClient(grpcConn),
		bankClient:   banktypes.NewQueryClient(grpcConn),
		encodeConfig: interfaceRegistry,
	}, nil
}
