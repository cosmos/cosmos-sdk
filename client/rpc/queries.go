package rpc

import (
	"context"

	gogogrpc "github.com/gogo/protobuf/grpc"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	"github.com/cosmos/cosmos-sdk/client"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	qtypes "github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/version"
)

// This is the struct that we will implement all the handlers on.
type queryServer struct {
	clientCtx         client.Context
	interfaceRegistry codectypes.InterfaceRegistry
}

var _ qtypes.QueryServer = queryServer{}

// NewQueryServer creates a new tendermint query server.
func NewQueryServer(clientCtx client.Context, interfaceRegistry codectypes.InterfaceRegistry) qtypes.QueryServer {
	return queryServer{
		clientCtx:         clientCtx,
		interfaceRegistry: interfaceRegistry,
	}
}

func (s queryServer) GetSyncing(context.Context, *qtypes.GetSyncingRequest) (*qtypes.GetSyncingResponse, error) {
	return &qtypes.GetSyncingResponse{}, nil
}

func (s queryServer) GetLatestBlock(context.Context, *qtypes.GetLatestBlockRequest) (*qtypes.GetLatestBlockResponse, error) {
	return &qtypes.GetLatestBlockResponse{}, nil
}

func (s queryServer) GetBlockByHeight(context.Context, *qtypes.GetBlockByHeightRequest) (*qtypes.GetBlockByHeightResponse, error) {
	return &qtypes.GetBlockByHeightResponse{}, nil
}

func (s queryServer) GetLatestValidatorSet(context.Context, *qtypes.GetLatestValidatorSetRequest) (*qtypes.GetLatestValidatorSetResponse, error) {
	return &qtypes.GetLatestValidatorSetResponse{}, nil
}

func (s queryServer) GetValidatorSetByHeight(context.Context, *qtypes.GetValidatorSetByHeightRequest) (*qtypes.GetValidatorSetByHeightResponse, error) {
	return &qtypes.GetValidatorSetByHeightResponse{}, nil
}

func (s queryServer) GetNodeInfo(ctx context.Context, req *qtypes.GetNodeInfoRequest) (*qtypes.GetNodeInfoResponse, error) {
	status, err := getNodeStatus(s.clientCtx)
	if err != nil {
		return nil, err
	}

	protoNodeInfo := status.NodeInfo.ToProto()
	nodeInfo := version.NewInfo()

	resp := qtypes.GetNodeInfoResponse{
		DefaultNodeInfo: protoNodeInfo,
		ApplicationVersion: &qtypes.VersionInfo{
			AppName:   nodeInfo.AppName,
			Name:      nodeInfo.Name,
			GitCommit: nodeInfo.GitCommit,
			GoVersion: nodeInfo.GoVersion,
			Version:   nodeInfo.Version,
			BuildTags: nodeInfo.BuildTags,
		},
	}
	return &resp, nil
}

// RegisterQueryService registers the tendermint queries on the gRPC router.
func RegisterQueryService(
	qrt gogogrpc.Server,
	clientCtx client.Context,
	interfaceRegistry codectypes.InterfaceRegistry,
) {
	qtypes.RegisterQueryServer(
		qrt,
		NewQueryServer(clientCtx, interfaceRegistry),
	)
}

// RegisterGRPCGatewayRoutes mounts the tendermint service's GRPC-gateway routes on the
// given Mux.
func RegisterGRPCGatewayRoutes(clientConn gogogrpc.ClientConn, mux *runtime.ServeMux) {
	qtypes.RegisterQueryHandlerClient(context.Background(), mux, qtypes.NewQueryClient(clientConn))
}
