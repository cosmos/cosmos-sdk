package sdk

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"

	"github.com/cosmos/cosmos-sdk/codec/types"
	auth "github.com/cosmos/cosmos-sdk/x/auth/types"

	"google.golang.org/grpc"
)

type Client struct {
	authClient   auth.QueryClient
	bankClient   banktypes.QueryClient
	encodeConfig types.InterfaceRegistry

	clientCtx client.Context

	endpoint string
}

// NewClient returns the client to call Cosmos RPC.
func NewClient(grpcEndpoint, tendermintRPCEndpoint string) (*Client, error) {
	// instantiate gRPC connection
	grpcConn, err := grpc.Dial(grpcEndpoint, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	// create interface registry, and register used modules types
	interfaceRegistry := types.NewInterfaceRegistry()
	auth.RegisterInterfaces(interfaceRegistry)
	banktypes.RegisterInterfaces(interfaceRegistry)
	cryptocodec.RegisterInterfaces(interfaceRegistry)

	// build client context for tendermint RPC
	// we cannot use .WithNodeURI because it panics on failures
	// so we can either do this or a recover call.
	rpc, err := rpchttp.New(tendermintRPCEndpoint, "/websocket")
	if err != nil {
		return nil, err
	}
	cdc := codec.NewProtoCodec(interfaceRegistry) // initialize a proto codec for the client context
	clientContext := client.Context{
		Client:  rpc,
		NodeURI: tendermintRPCEndpoint,
	}.
		WithJSONMarshaler(cdc).
		WithInterfaceRegistry(interfaceRegistry).
		WithTxConfig(tx.NewTxConfig(cdc, tx.DefaultSignModes)).
		WithAccountRetriever(auth.AccountRetriever{}).
		WithBroadcastMode(flags.BroadcastBlock)

	return &Client{
		authClient:   auth.NewQueryClient(grpcConn),
		bankClient:   banktypes.NewQueryClient(grpcConn),
		encodeConfig: interfaceRegistry,
		clientCtx:    clientContext,
	}, nil
}
