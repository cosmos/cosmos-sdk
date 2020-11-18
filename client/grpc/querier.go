package tmservice

import (
	"context"

	gogogrpc "github.com/gogo/protobuf/grpc"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	qtypes "github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/version"
)

// This is the struct that we will implement all the handlers on.
type queryServer struct {
	clientCtx         client.Context
	interfaceRegistry codectypes.InterfaceRegistry
}

var _ qtypes.ServiceServer = queryServer{}

// NewQueryServer creates a new tendermint query server.
func NewQueryServer(clientCtx client.Context, interfaceRegistry codectypes.InterfaceRegistry) qtypes.ServiceServer {
	return queryServer{
		clientCtx:         clientCtx,
		interfaceRegistry: interfaceRegistry,
	}
}

func (s queryServer) GetSyncing(context.Context, *qtypes.GetSyncingRequest) (*qtypes.GetSyncingResponse, error) {
	status, err := getNodeStatus(s.clientCtx)
	if err != nil {
		return nil, err
	}
	return &qtypes.GetSyncingResponse{
		Syncing: status.SyncInfo.CatchingUp,
	}, nil
}

func (s queryServer) GetLatestBlock(context.Context, *qtypes.GetLatestBlockRequest) (*qtypes.GetLatestBlockResponse, error) {
	status, err := getBlock(s.clientCtx, nil)
	if err != nil {
		return nil, err
	}
	protoBlockId := status.BlockID.ToProto()
	protoBlock, err := status.Block.ToProto()
	if err != nil {
		return nil, err
	}
	return &qtypes.GetLatestBlockResponse{
		BlockId: &protoBlockId,
		Block:   protoBlock,
	}, nil
}

func (s queryServer) GetBlockByHeight(_ context.Context, req *qtypes.GetBlockByHeightRequest) (*qtypes.GetBlockByHeightResponse, error) {
	chainHeight, err := rpc.GetChainHeight(s.clientCtx)
	if err != nil {
		return nil, err
	}

	if req.Height > chainHeight {
		return nil, status.Error(codes.InvalidArgument, "requested block height is bigger then the chain length")
	}

	res, err := getBlock(s.clientCtx, &req.Height)
	if err != nil {
		return nil, err
	}
	protoBlockId := res.BlockID.ToProto()
	protoBlock, err := res.Block.ToProto()
	if err != nil {
		return nil, err
	}
	return &qtypes.GetBlockByHeightResponse{
		BlockId: &protoBlockId,
		Block:   protoBlock,
	}, nil
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

// RegisterTendermintService registers the tendermint queries on the gRPC router.
func RegisterTendermintService(
	qrt gogogrpc.Server,
	clientCtx client.Context,
	interfaceRegistry codectypes.InterfaceRegistry,
) {
	qtypes.RegisterServiceServer(
		qrt,
		NewQueryServer(clientCtx, interfaceRegistry),
	)
}

// RegisterGRPCGatewayRoutes mounts the tendermint service's GRPC-gateway routes on the
// given Mux.
func RegisterGRPCGatewayRoutes(clientConn gogogrpc.ClientConn, mux *runtime.ServeMux) {
	qtypes.RegisterServiceHandlerClient(context.Background(), mux, qtypes.NewServiceClient(clientConn))
}
