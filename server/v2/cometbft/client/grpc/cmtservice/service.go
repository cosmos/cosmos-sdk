package cmtservice

import (
	"context"
	"strings"

	abci "github.com/cometbft/cometbft/api/cometbft/abci/v1"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	gogogrpc "github.com/cosmos/gogoproto/grpc"
	gogoprotoany "github.com/cosmos/gogoproto/types/any"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"cosmossdk.io/server/v2/cometbft/client/rpc"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	qtypes "github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/version"
)

var (
	_ ServiceServer                        = queryServer{}
	_ gogoprotoany.UnpackInterfacesMessage = &GetLatestValidatorSetResponse{}
)

const (
	QueryPathApp   = "app"
	QueryPathP2P   = "p2p"
	QueryPathStore = "store"
)

type (
	abciQueryFn = func(context.Context, *abci.QueryRequest) (*abci.QueryResponse, error)

	queryServer struct {
		client  rpc.CometRPC
		queryFn abciQueryFn
	}
)

// NewQueryServer creates a new CometBFT query server.
func NewQueryServer(
	client rpc.CometRPC,
	queryFn abciQueryFn,
) ServiceServer {
	return queryServer{
		queryFn: queryFn,
		client:  client,
	}
}

// GetNodeStatus returns the status of the node.
func (s queryServer) GetNodeStatus(ctx context.Context) (*coretypes.ResultStatus, error) {
	return s.client.Status(ctx)
}

// GetSyncing implements ServiceServer.GetSyncing
func (s queryServer) GetSyncing(ctx context.Context, _ *GetSyncingRequest) (*GetSyncingResponse, error) {
	status, err := s.client.Status(ctx)
	if err != nil {
		return nil, err
	}

	return &GetSyncingResponse{
		Syncing: status.SyncInfo.CatchingUp,
	}, nil
}

// GetLatestBlock implements ServiceServer.GetLatestBlock
func (s queryServer) GetLatestBlock(ctx context.Context, _ *GetLatestBlockRequest) (*GetLatestBlockResponse, error) {
	block, err := s.client.Block(ctx, nil)
	if err != nil {
		return nil, err
	}

	protoBlockID := block.BlockID.ToProto()
	protoBlock, err := block.Block.ToProto()
	if err != nil {
		return nil, err
	}

	return &GetLatestBlockResponse{
		BlockId:  &protoBlockID,
		Block:    protoBlock,
		SdkBlock: convertBlock(protoBlock),
	}, nil
}

// GetBlockByHeight implements ServiceServer.GetBlockByHeight
func (s queryServer) GetBlockByHeight(
	ctx context.Context,
	req *GetBlockByHeightRequest,
) (*GetBlockByHeightResponse, error) {
	nodeStatus, err := s.client.Status(ctx)
	if err != nil {
		return nil, err
	}

	blockHeight := nodeStatus.SyncInfo.LatestBlockHeight

	if req.Height > blockHeight {
		return nil, status.Error(codes.InvalidArgument, "requested block height is bigger then the chain length")
	}

	b, err := s.client.Block(ctx, &req.Height)
	if err != nil {
		return nil, err
	}

	protoBlockID := b.BlockID.ToProto()
	protoBlock, err := b.Block.ToProto()
	if err != nil {
		return nil, err
	}

	return &GetBlockByHeightResponse{
		BlockId:  &protoBlockID,
		Block:    protoBlock,
		SdkBlock: convertBlock(protoBlock),
	}, nil
}

// GetLatestValidatorSet implements ServiceServer.GetLatestValidatorSet
func (s queryServer) GetLatestValidatorSet(
	ctx context.Context,
	req *GetLatestValidatorSetRequest,
) (*GetLatestValidatorSetResponse, error) {
	page, limit, err := qtypes.ParsePagination(req.Pagination)
	if err != nil {
		return nil, err
	}

	return ValidatorsOutput(ctx, s.client, nil, page, limit)
}

func (m *GetLatestValidatorSetResponse) UnpackInterfaces(unpacker gogoprotoany.AnyUnpacker) error {
	var pubKey cryptotypes.PubKey
	for _, val := range m.Validators {
		err := unpacker.UnpackAny(val.PubKey, &pubKey)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetValidatorSetByHeight implements ServiceServer.GetValidatorSetByHeight
func (s queryServer) GetValidatorSetByHeight(
	ctx context.Context,
	req *GetValidatorSetByHeightRequest,
) (*GetValidatorSetByHeightResponse, error) {
	page, limit, err := qtypes.ParsePagination(req.Pagination)
	if err != nil {
		return nil, err
	}

	nodeStatus, err := s.client.Status(ctx)
	if err != nil {
		return nil, err
	}

	blockHeight := nodeStatus.SyncInfo.LatestBlockHeight

	if req.Height > blockHeight {
		return nil, status.Error(codes.InvalidArgument, "requested block height is bigger then the chain length")
	}

	r, err := ValidatorsOutput(ctx, s.client, &req.Height, page, limit)
	if err != nil {
		return nil, err
	}

	return &GetValidatorSetByHeightResponse{
		BlockHeight: r.BlockHeight,
		Validators:  r.Validators,
		Pagination:  r.Pagination,
	}, nil
}

func ValidatorsOutput(
	ctx context.Context,
	client rpc.CometRPC,
	height *int64,
	page, limit int,
) (*GetLatestValidatorSetResponse, error) {
	vs, err := client.Validators(ctx, height, &page, &limit)
	if err != nil {
		return nil, err
	}

	resp := GetLatestValidatorSetResponse{
		BlockHeight: vs.BlockHeight,
		Validators:  make([]*Validator, len(vs.Validators)),
		Pagination: &qtypes.PageResponse{
			Total: uint64(vs.Total),
		},
	}

	for i, v := range vs.Validators {
		pk, err := cryptocodec.FromCmtPubKeyInterface(v.PubKey)
		if err != nil {
			return nil, err
		}
		anyPub, err := codectypes.NewAnyWithValue(pk)
		if err != nil {
			return nil, err
		}

		resp.Validators[i] = &Validator{
			Address:          sdk.ConsAddress(v.Address).String(),
			ProposerPriority: v.ProposerPriority,
			PubKey:           anyPub,
			VotingPower:      v.VotingPower,
		}
	}

	return &resp, nil
}

// GetNodeInfo implements ServiceServer.GetNodeInfo
func (s queryServer) GetNodeInfo(ctx context.Context, _ *GetNodeInfoRequest) (*GetNodeInfoResponse, error) {
	nodeStatus, err := s.client.Status(ctx)
	if err != nil {
		return nil, err
	}

	protoNodeInfo := nodeStatus.NodeInfo.ToProto()
	nodeInfo := version.NewInfo()

	deps := make([]*Module, len(nodeInfo.BuildDeps))

	for i, dep := range nodeInfo.BuildDeps {
		deps[i] = &Module{
			Path:    dep.Path,
			Sum:     dep.Sum,
			Version: dep.Version,
		}
	}

	resp := GetNodeInfoResponse{
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
	return &resp, nil
}

func (s queryServer) ABCIQuery(ctx context.Context, req *ABCIQueryRequest) (*ABCIQueryResponse, error) {
	if s.queryFn == nil {
		return nil, status.Error(codes.Internal, "ABCI Query handler undefined")
	}
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if len(req.Path) == 0 {
		return nil, status.Error(codes.InvalidArgument, "empty query path")
	}

	if path := SplitABCIQueryPath(req.Path); len(path) > 0 {
		switch path[0] {
		case QueryPathApp, QueryPathStore, QueryPathP2P:
			// valid path

		default:
			// Otherwise, error as to prevent either valid gRPC service requests or
			// bogus ABCI queries.
			return nil, status.Errorf(codes.InvalidArgument, "unsupported ABCI query path: %s", req.Path)
		}
	}

	res, err := s.queryFn(ctx, req.ToABCIRequestQuery())
	if err != nil {
		return nil, err
	}
	return FromABCIResponseQuery(res), nil
}

// RegisterTendermintService registers the CometBFT queries on the gRPC router.
func RegisterTendermintService(
	client rpc.CometRPC,
	server gogogrpc.Server,
	queryFn abciQueryFn,
) {
	RegisterServiceServer(server, NewQueryServer(client, queryFn))
}

// RegisterGRPCGatewayRoutes mounts the CometBFT service's GRPC-gateway routes on the
// given Mux.
func RegisterGRPCGatewayRoutes(clientConn gogogrpc.ClientConn, mux *runtime.ServeMux) {
	_ = RegisterServiceHandlerClient(context.Background(), mux, NewServiceClient(clientConn))
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
