package cmtservice

import (
	"context"
	"strings"

	abci "github.com/cometbft/cometbft/api/cometbft/abci/v1"
	gogogrpc "github.com/cosmos/gogoproto/grpc"
	gogoprotoany "github.com/cosmos/gogoproto/types/any"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	"cosmossdk.io/core/address"
	"cosmossdk.io/server/v2/cometbft/client/rpc"

	cmtservice "github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
)

var _ gogoprotoany.UnpackInterfacesMessage = &cmtservice.GetLatestValidatorSetResponse{}

const (
	QueryPathApp   = "app"
	QueryPathP2P   = "p2p"
	QueryPathStore = "store"
)

type abciQueryFn = func(context.Context, *abci.QueryRequest) (*abci.QueryResponse, error)

// RegisterTendermintService registers the CometBFT queries on the gRPC router.
func RegisterTendermintService(
	client rpc.CometRPC,
	server gogogrpc.Server,
	queryFn abciQueryFn,
	consensusCodec address.Codec,
) {
	cmtservice.RegisterServiceServer(server, cmtservice.NewQueryServer(client, queryFn, consensusCodec))
}

// RegisterGRPCGatewayRoutes mounts the CometBFT service's GRPC-gateway routes on the
// given Mux.
func RegisterGRPCGatewayRoutes(clientConn gogogrpc.ClientConn, mux *runtime.ServeMux) {
	_ = cmtservice.RegisterServiceHandlerClient(context.Background(), mux, cmtservice.NewServiceClient(clientConn))
}

// SplitABCIQueryPath splits a string path using the delimiter '/'.
//
// e.g. "this/is/funny" becomes []string{"this", "is", "funny"}
func SplitABCIQueryPath(requestPath string) (path []string) {
	path = strings.Split(requestPath, "/")

	// first element is empty string
	if len(path) > 0 && path[0] == "" {
		path = path[1:]
	}

	return path
}
