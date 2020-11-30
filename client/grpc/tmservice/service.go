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

// GetSyncing implements ServiceServer.GetSyncing
func (s queryServer) GetSyncing(_ context.Context, _ *qtypes.GetSyncingRequest) (*qtypes.GetSyncingResponse, error) {
	status, err := getNodeStatus(s.clientCtx)
	if err != nil {
		return nil, err
	}
	return &qtypes.GetSyncingResponse{
		Syncing: status.SyncInfo.CatchingUp,
	}, nil
}

// GetLatestBlock implements ServiceServer.GetLatestBlock
func (s queryServer) GetLatestBlock(context.Context, *qtypes.GetLatestBlockRequest) (*qtypes.GetLatestBlockResponse, error) {
	status, err := getBlock(s.clientCtx, nil)
	if err != nil {
		return nil, err
	}

	protoBlockID := status.BlockID.ToProto()
	protoBlock, err := status.Block.ToProto()
	if err != nil {
		return nil, err
	}

	return &qtypes.GetLatestBlockResponse{
		BlockId: &protoBlockID,
		Block:   protoBlock,
	}, nil
}

// GetBlockByHeight implements ServiceServer.GetBlockByHeight
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
	protoBlockID := res.BlockID.ToProto()
	protoBlock, err := res.Block.ToProto()
	if err != nil {
		return nil, err
	}
	return &qtypes.GetBlockByHeightResponse{
		BlockId: &protoBlockID,
		Block:   protoBlock,
	}, nil
}

// GetLatestValidatorSet implements ServiceServer.GetLatestValidatorSet
func (s queryServer) GetLatestValidatorSet(ctx context.Context, req *qtypes.GetLatestValidatorSetRequest) (*qtypes.GetLatestValidatorSetResponse, error) {
	page, limit, err := qtypes.ParsePagination(req.Pagination)
	if err != nil {
		return nil, err
	}

	validatorsRes, err := rpc.GetValidators(s.clientCtx, nil, &page, &limit)
	if err != nil {
		return nil, err
	}

	outputValidatorsRes := &qtypes.GetLatestValidatorSetResponse{
		BlockHeight: validatorsRes.BlockHeight,
		Validators:  make([]*qtypes.Validator, len(validatorsRes.Validators)),
	}

	for i, validator := range validatorsRes.Validators {
		outputValidatorsRes.Validators[i] = &qtypes.Validator{
			Address:          validator.Address,
			ProposerPriority: validator.ProposerPriority,
			PubKey:           validator.PubKey,
			VotingPower:      validator.VotingPower,
		}
	}
	return outputValidatorsRes, nil
}

// GetValidatorSetByHeight implements ServiceServer.GetValidatorSetByHeight
func (s queryServer) GetValidatorSetByHeight(ctx context.Context, req *qtypes.GetValidatorSetByHeightRequest) (*qtypes.GetValidatorSetByHeightResponse, error) {
	page, limit, err := qtypes.ParsePagination(req.Pagination)
	if err != nil {
		return nil, err
	}

	chainHeight, err := rpc.GetChainHeight(s.clientCtx)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to parse chain height")
	}
	if req.Height > chainHeight {
		return nil, status.Error(codes.InvalidArgument, "requested block height is bigger then the chain length")
	}

	validatorsRes, err := rpc.GetValidators(s.clientCtx, &req.Height, &page, &limit)

	if err != nil {
		return nil, err
	}

	outputValidatorsRes := &qtypes.GetValidatorSetByHeightResponse{
		BlockHeight: validatorsRes.BlockHeight,
		Validators:  make([]*qtypes.Validator, len(validatorsRes.Validators)),
	}

	for i, validator := range validatorsRes.Validators {
		outputValidatorsRes.Validators[i] = &qtypes.Validator{
			Address:          validator.Address,
			ProposerPriority: validator.ProposerPriority,
			PubKey:           validator.PubKey,
			VotingPower:      validator.VotingPower,
		}
	}
	return outputValidatorsRes, nil
}

// GetNodeInfo implements ServiceServer.GetNodeInfo
func (s queryServer) GetNodeInfo(ctx context.Context, req *qtypes.GetNodeInfoRequest) (*qtypes.GetNodeInfoResponse, error) {
	status, err := getNodeStatus(s.clientCtx)
	if err != nil {
		return nil, err
	}

	protoNodeInfo := status.NodeInfo.ToProto()
	nodeInfo := version.NewInfo()

	deps := make([]*qtypes.Module, len(nodeInfo.BuildDeps))

	for i, dep := range nodeInfo.BuildDeps {
		deps[i] = &qtypes.Module{
			Path:    dep.Path,
			Sum:     dep.Sum,
			Version: dep.Version,
		}
	}

	resp := qtypes.GetNodeInfoResponse{
		DefaultNodeInfo: protoNodeInfo,
		ApplicationVersion: &qtypes.VersionInfo{
			AppName:   nodeInfo.AppName,
			Name:      nodeInfo.Name,
			GitCommit: nodeInfo.GitCommit,
			GoVersion: nodeInfo.GoVersion,
			Version:   nodeInfo.Version,
			BuildTags: nodeInfo.BuildTags,
			BuildDeps: deps,
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
