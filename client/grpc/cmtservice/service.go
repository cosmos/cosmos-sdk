package cmtservice

import (
	"context"
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"
	gogogrpc "github.com/cosmos/gogoproto/grpc"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	qtypes "github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/version"
)

var (
	_ ServiceServer                      = queryServer{}
	_ codectypes.UnpackInterfacesMessage = &GetLatestValidatorSetResponse{}
)

type (
	abciQueryFn = func(context.Context, *abci.RequestQuery) (*abci.ResponseQuery, error)

	queryServer struct {
		clientCtx         client.Context
		interfaceRegistry codectypes.InterfaceRegistry
		queryFn           abciQueryFn
	}
)

func NewQueryServer(
	clientCtx client.Context,
	interfaceRegistry codectypes.InterfaceRegistry,
	queryFn abciQueryFn,
) ServiceServer {
	return queryServer{
		clientCtx:         clientCtx,
		interfaceRegistry: interfaceRegistry,
		queryFn:           queryFn,
	}
}

// GetSyncing implements ServiceServer.GetSyncing
func (s queryServer) GetSyncing(ctx context.Context, _ *GetSyncingRequest) (*GetSyncingResponse, error) {
	status, err := GetNodeStatus(ctx, s.clientCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to get node status: %w", err)
	}
	return &GetSyncingResponse{
		Syncing: status.SyncInfo.CatchingUp,
	}, nil
}

// GetLatestBlock implements ServiceServer.GetLatestBlock
func (s queryServer) GetLatestBlock(ctx context.Context, _ *GetLatestBlockRequest) (*GetLatestBlockResponse, error) {
	return s.getBlockByHeight(ctx, nil)
}

// GetBlockByHeight implements ServiceServer.GetBlockByHeight
func (s queryServer) GetBlockByHeight(ctx context.Context, req *GetBlockByHeightRequest) (*GetBlockByHeightResponse, error) {
	blockHeight, err := getBlockHeight(ctx, s.clientCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to get block height: %w", err)
	}

	if req.Height > blockHeight {
		return nil, status.Error(codes.InvalidArgument, "requested block height is greater than the chain length")
	}

	return s.getBlockByHeight(ctx, &req.Height)
}

func (s queryServer) getBlockByHeight(ctx context.Context, height *int64) (*GetBlockByHeightResponse, error) {
	status, err := getBlock(ctx, s.clientCtx, height)
	if err != nil {
		return nil, fmt.Errorf("failed to get block: %w", err)
	}

	protoBlockID := status.BlockID.ToProto()
	protoBlock, err := status.Block.ToProto()
	if err != nil {
		return nil, fmt.Errorf("failed to convert block to proto: %w", err)
	}

	return &GetBlockByHeightResponse{
		BlockId:  &protoBlockID,
		Block:    protoBlock,
		SdkBlock: convertBlock(protoBlock),
	}, nil
}

// GetLatestValidatorSet implements ServiceServer.GetLatestValidatorSet
func (s queryServer) GetLatestValidatorSet(ctx context.Context, req *GetLatestValidatorSetRequest) (*GetLatestValidatorSetResponse, error) {
	page, limit, err := qtypes.ParsePagination(req.Pagination)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pagination: %w", err)
	}

	return s.getValidators(ctx, nil, page, limit)
}

// GetValidatorSetByHeight implements ServiceServer.GetValidatorSetByHeight
func (s queryServer) GetValidatorSetByHeight(ctx context.Context, req *GetValidatorSetByHeightRequest) (*GetValidatorSetByHeightResponse, error) {
	page, limit, err := qtypes.ParsePagination(req.Pagination)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pagination: %w", err)
	}

	blockHeight, err := getBlockHeight(ctx, s.clientCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to get block height: %w", err)
	}

	if req.Height > blockHeight {
		return nil, status.Error(codes.InvalidArgument, "requested block height is greater than the chain length")
	}

	r, err := s.getValidators(ctx, &req.Height, page, limit)
	if err != nil {
		return nil, err
	}

	return &GetValidatorSetByHeightResponse{
		BlockHeight: r.BlockHeight,
		Validators:  r.Validators,
		Pagination:  r.Pagination,
	}, nil
}

func (s queryServer) getValidators(ctx context.Context, height *int64, page, limit int) (*GetLatestValidatorSetResponse, error) {
	vs, err := getValidators(ctx, s.clientCtx, height, page, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get validators: %w", err)
	}

	resp := &GetLatestValidatorSetResponse{
		BlockHeight: vs.BlockHeight,
		Validators:  make([]*Validator, len(vs.Validators)),
		Pagination: &qtypes.PageResponse{
			Total: uint64(vs.Total),
		},
	}

	for i, v := range vs.Validators {
		pk, err := cryptocodec.FromCmtPubKeyInterface(v.PubKey)
		if err != nil {
			return nil, fmt.Errorf("failed to decode public key: %w", err)
		}
		anyPub, err := codectypes.NewAnyWithValue(pk)
		if err != nil {
			return nil, fmt.Errorf("failed to create Any type with public key: %w", err)
		}

		resp.Validators[i] = &Validator{
			Address:          sdk.ConsAddress(v.Address).String(),
			ProposerPriority: v.ProposerPriority,
			PubKey:           anyPub,
			VotingPower:      v.VotingPower,
		}
	}

	return resp, nil
}

// GetNodeInfo implements ServiceServer.GetNodeInfo
func (s queryServer) GetNodeInfo(ctx context.Context, _ *GetNodeInfoRequest) (*GetNodeInfoResponse, error) {
	status, err := GetNodeStatus(ctx, s.clientCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to get node status: %w", err)
	}

	protoNodeInfo := status.NodeInfo.ToProto()
	nodeInfo := version.NewInfo()

	deps := make([]*Module, len(nodeInfo.BuildDeps))
	for i, dep := range nodeInfo.BuildDeps {
		deps[i] = &Module{
			Path:    dep.Path,
			Sum:     dep.Sum,
			Version: dep.Version,
		}
	}

	resp := &GetNodeInfoResponse{
		DefaultNodeInfo: protoNodeInfo,
		ApplicationVersion: &VersionInfo{
			AppName:          nodeInfo.AppName,
			Name:             nodeInfo.Name,
			GitCommit:        nodeInfo.GitCommit,
			GoVersion:        nodeInfo.GoVersion,
			Version:          nodeInfo.Version,
			BuildTags:        nodeInfo.BuildTags,
			BuildDeps:        deps,
			CosmosSdkVersion: nodeInfo.CosmosSdkVersion,
		},
	}
	return resp, nil
}

// ABCIQuery implements ServiceServer.ABCIQuery
func (s queryServer) ABCIQuery(ctx context.Context, req *ABCIQueryRequest) (*ABCIQueryResponse, error) {
	if s.queryFn == nil {
		return nil, status.Error(codes.Internal, "ABCI Query handler undefined")
	}
	if req == nil || len(req.Path) == 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid request path")
	}

	if path := baseapp.SplitABCIQueryPath(req.Path); len(path) > 0 {
		switch path[0] {
		case baseapp.QueryPathApp, baseapp.QueryPathStore, baseapp.QueryPathP2P, baseapp.QueryPathCustom:
		default:
			return nil, status.Errorf(codes.InvalidArgument, "unsupported ABCI query path: %s", req.Path)
		}
	}

	res, err := s.queryFn(ctx, req.ToABCIRequestQuery())
	if err != nil {
		return nil, fmt.Errorf("ABCI query failed: %w", err)
	}
	return FromABCIResponseQuery(res), nil
}

func RegisterTendermintService(
	clientCtx client.Context,
	server gogogrpc.Server,
	iRegistry codectypes.InterfaceRegistry,
	queryFn abciQueryFn,
) {
	RegisterServiceServer(server, NewQueryServer(clientCtx, iRegistry, queryFn))
}

func RegisterGRPCGatewayRoutes(clientConn gogogrpc.ClientConn, mux *runtime.ServeMux) {
	_ = RegisterServiceHandlerClient(context.Background(), mux, NewServiceClient(clientConn))
}
