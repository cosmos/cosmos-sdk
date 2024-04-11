package cmtservice

import (
	"context"
	"strings"

	tmp2p "buf.build/gen/go/tendermint/tendermint/protocolbuffers/go/tendermint/p2p"
	tmtypes "buf.build/gen/go/tendermint/tendermint/protocolbuffers/go/tendermint/types"
	queryv1beta1 "cosmossdk.io/api/cosmos/base/query/v1beta1"
	tmv1beta1 "cosmossdk.io/api/cosmos/base/tendermint/v1beta1"
	"cosmossdk.io/core/address"
	"cosmossdk.io/server/v2/cometbft/client/rpc"
	abci "github.com/cometbft/cometbft/abci/types"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	gogogrpc "github.com/cosmos/gogoproto/grpc"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/version"
)

var (
	_ tmv1beta1.ServiceServer = queryServer{}
)

type (
	abciQueryFn = func(context.Context, *abci.RequestQuery) (*abci.ResponseQuery, error)

	queryServer struct {
		tmv1beta1.UnimplementedServiceServer
		client      rpc.CometRPC
		queryFn     abciQueryFn
		consAddrCdc address.Codec
	}
)

// NewQueryServer creates a new CometBFT query server.
func NewQueryServer(
	client rpc.CometRPC,
	queryFn abciQueryFn,
	consAddrCdc address.Codec,
) tmv1beta1.ServiceServer {
	return &queryServer{
		queryFn:     queryFn,
		client:      client,
		consAddrCdc: consAddrCdc,
	}
}

// GetNodeStatus returns the status of the node.
func (s queryServer) GetNodeStatus(ctx context.Context) (*coretypes.ResultStatus, error) {
	return s.client.Status(ctx)
}

// GetSyncing implements ServiceServer.GetSyncing
func (s queryServer) GetSyncing(ctx context.Context, _ *tmv1beta1.GetSyncingRequest) (*tmv1beta1.GetSyncingResponse, error) {
	status, err := s.client.Status(ctx)
	if err != nil {
		return nil, err
	}

	return &tmv1beta1.GetSyncingResponse{
		Syncing: status.SyncInfo.CatchingUp,
	}, nil
}

// GetLatestBlock implements ServiceServer.GetLatestBlock
func (s queryServer) GetLatestBlock(ctx context.Context, _ *tmv1beta1.GetLatestBlockRequest) (*tmv1beta1.GetLatestBlockResponse, error) {
	block, err := s.client.Block(ctx, nil)
	if err != nil {
		return nil, err
	}

	return &tmv1beta1.GetLatestBlockResponse{
		BlockId: &tmtypes.BlockID{
			Hash: block.BlockID.Hash,
			PartSetHeader: &tmtypes.PartSetHeader{
				Total: block.BlockID.PartSetHeader.Total,
				Hash:  block.BlockID.PartSetHeader.Hash,
			},
		},
		Block:    blockToProto(block.Block),
		SdkBlock: blockToSdkBlock(block.Block, s.consAddrCdc),
	}, nil
}

// GetBlockByHeight implements ServiceServer.GetBlockByHeight
func (s queryServer) GetBlockByHeight(ctx context.Context, req *tmv1beta1.GetBlockByHeightRequest) (*tmv1beta1.GetBlockByHeightResponse, error) {
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

	return &tmv1beta1.GetBlockByHeightResponse{
		BlockId: &tmtypes.BlockID{
			Hash: b.BlockID.Hash,
			PartSetHeader: &tmtypes.PartSetHeader{
				Total: b.BlockID.PartSetHeader.Total,
				Hash:  b.BlockID.PartSetHeader.Hash,
			},
		},
		Block:    blockToProto(b.Block),
		SdkBlock: blockToSdkBlock(b.Block, s.consAddrCdc),
	}, nil
}

// GetLatestValidatorSet implements ServiceServer.GetLatestValidatorSet
func (s queryServer) GetLatestValidatorSet(ctx context.Context, req *tmv1beta1.GetLatestValidatorSetRequest) (*tmv1beta1.GetLatestValidatorSetResponse, error) {
	page, limit, err := ParsePagination(req.Pagination)
	if err != nil {
		return nil, err
	}

	return ValidatorsOutput(ctx, s.consAddrCdc, s.client, nil, page, limit)
}

// GetValidatorSetByHeight implements ServiceServer.GetValidatorSetByHeight
func (s queryServer) GetValidatorSetByHeight(ctx context.Context, req *tmv1beta1.GetValidatorSetByHeightRequest) (*tmv1beta1.GetValidatorSetByHeightResponse, error) {
	page, limit, err := ParsePagination(req.Pagination)
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

	r, err := ValidatorsOutput(ctx, s.consAddrCdc, s.client, &req.Height, page, limit)
	if err != nil {
		return nil, err
	}

	return &tmv1beta1.GetValidatorSetByHeightResponse{
		BlockHeight: r.BlockHeight,
		Validators:  r.Validators,
		Pagination:  r.Pagination,
	}, nil
}

func ValidatorsOutput(ctx context.Context, consAddrCdc address.Codec, client rpc.CometRPC, height *int64, page, limit int) (*tmv1beta1.GetLatestValidatorSetResponse, error) {
	vs, err := client.Validators(ctx, height, &page, &limit)
	if err != nil {
		return nil, err
	}

	resp := &tmv1beta1.GetLatestValidatorSetResponse{
		BlockHeight: vs.BlockHeight,
		Validators:  make([]*tmv1beta1.Validator, len(vs.Validators)),
		Pagination: &queryv1beta1.PageResponse{
			Total: uint64(vs.Total),
		},
	}

	for i, v := range vs.Validators {
		pk, err := pubKeyToProto(v.PubKey)
		if err != nil {
			return nil, err
		}

		addr, err := consAddrCdc.BytesToString(v.Address.Bytes())
		if err != nil {
			return nil, err
		}

		resp.Validators[i] = &tmv1beta1.Validator{
			Address:          addr,
			ProposerPriority: v.ProposerPriority,
			PubKey:           pk,
			VotingPower:      v.VotingPower,
		}
	}

	return resp, nil
}

// GetNodeInfo implements ServiceServer.GetNodeInfo
func (s queryServer) GetNodeInfo(ctx context.Context, _ *tmv1beta1.GetNodeInfoRequest) (*tmv1beta1.GetNodeInfoResponse, error) {
	nodeStatus, err := s.client.Status(ctx)
	if err != nil {
		return nil, err
	}

	nodeInfo := version.NewInfo()

	deps := make([]*tmv1beta1.Module, len(nodeInfo.BuildDeps))

	for i, dep := range nodeInfo.BuildDeps {
		deps[i] = &tmv1beta1.Module{
			Path:    dep.Path,
			Sum:     dep.Sum,
			Version: dep.Version,
		}
	}

	resp := tmv1beta1.GetNodeInfoResponse{
		DefaultNodeInfo: &tmp2p.DefaultNodeInfo{
			ProtocolVersion: &tmp2p.ProtocolVersion{
				P2P:   nodeStatus.NodeInfo.ProtocolVersion.P2P,
				Block: nodeStatus.NodeInfo.ProtocolVersion.Block,
				App:   nodeStatus.NodeInfo.ProtocolVersion.App,
			},
			DefaultNodeId: string(nodeStatus.NodeInfo.DefaultNodeID),
			ListenAddr:    nodeStatus.NodeInfo.ListenAddr,
			Network:       nodeStatus.NodeInfo.Network,
			Version:       nodeStatus.NodeInfo.Version,
			Channels:      nodeStatus.NodeInfo.Channels,
			Moniker:       nodeStatus.NodeInfo.Moniker,
			Other: &tmp2p.DefaultNodeInfoOther{
				TxIndex:    nodeStatus.NodeInfo.Other.TxIndex,
				RpcAddress: nodeStatus.NodeInfo.Other.RPCAddress,
			},
		},
		ApplicationVersion: &tmv1beta1.VersionInfo{
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

func (s queryServer) ABCIQuery(ctx context.Context, req *tmv1beta1.ABCIQueryRequest) (*tmv1beta1.ABCIQueryResponse, error) {
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
		case "app", "store", "p2p", "custom":
			// valid path
			// TODO complete this

		default:
			// Otherwise, error as to prevent either valid gRPC service requests or
			// bogus ABCI queries.
			return nil, status.Errorf(codes.InvalidArgument, "unsupported ABCI query path: %s", req.Path)
		}
	}

	res, err := s.queryFn(ctx, ToABCIRequestQuery(req))
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
	consAddrCodec address.Codec,
) {
	tmv1beta1.RegisterServiceServer(server, NewQueryServer(client, queryFn, consAddrCodec))
}

// RegisterGRPCGatewayRoutes mounts the CometBFT service's GRPC-gateway routes on the
// given Mux.
func RegisterGRPCGatewayRoutes(clientConn gogogrpc.ClientConn, mux *runtime.ServeMux) {
	// TODO: how to fix this? the generated gw file uses gogoproto.
	_ = RegisterServiceHandlerClient(context.Background(), mux, tmv1beta1.NewServiceClient(clientConn))
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
