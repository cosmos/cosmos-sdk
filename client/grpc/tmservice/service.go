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
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	qtypes "github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/version"
)

// This is the struct that we will implement all the handlers on.
type queryServer struct {
	clientCtx         client.Context
	interfaceRegistry codectypes.InterfaceRegistry
}

var _ ServiceServer = queryServer{}
var _ codectypes.UnpackInterfacesMessage = &GetLatestValidatorSetResponse{}

// NewQueryServer creates a new tendermint query server.
func NewQueryServer(clientCtx client.Context, interfaceRegistry codectypes.InterfaceRegistry) ServiceServer {
	return queryServer{
		clientCtx:         clientCtx,
		interfaceRegistry: interfaceRegistry,
	}
}

// GetSyncing implements ServiceServer.GetSyncing
func (s queryServer) GetSyncing(_ context.Context, _ *GetSyncingRequest) (*GetSyncingResponse, error) {
	status, err := getNodeStatus(s.clientCtx)
	if err != nil {
		return nil, err
	}
	return &GetSyncingResponse{
		Syncing: status.SyncInfo.CatchingUp,
	}, nil
}

// GetLatestBlock implements ServiceServer.GetLatestBlock
func (s queryServer) GetLatestBlock(context.Context, *GetLatestBlockRequest) (*GetLatestBlockResponse, error) {
	status, err := getBlock(s.clientCtx, nil)
	if err != nil {
		return nil, err
	}

	protoBlockID := status.BlockID.ToProto()
	protoBlock, err := status.Block.ToProto()
	if err != nil {
		return nil, err
	}

	return &GetLatestBlockResponse{
		BlockId: &protoBlockID,
		Block:   protoBlock,
	}, nil
}

// GetBlockByHeight implements ServiceServer.GetBlockByHeight
func (s queryServer) GetBlockByHeight(_ context.Context, req *GetBlockByHeightRequest) (*GetBlockByHeightResponse, error) {
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
	return &GetBlockByHeightResponse{
		BlockId: &protoBlockID,
		Block:   protoBlock,
	}, nil
}

// GetLatestValidatorSet implements ServiceServer.GetLatestValidatorSet
func (s queryServer) GetLatestValidatorSet(ctx context.Context, req *GetLatestValidatorSetRequest) (*GetLatestValidatorSetResponse, error) {
	page, limit, err := qtypes.ParsePagination(req.Pagination)
	if err != nil {
		return nil, err
	}

	validatorsRes, err := rpc.GetValidators(s.clientCtx, nil, &page, &limit)
	if err != nil {
		return nil, err
	}

	outputValidatorsRes := &GetLatestValidatorSetResponse{
		BlockHeight: validatorsRes.BlockHeight,
		Validators:  make([]*Validator, len(validatorsRes.Validators)),
	}

	for i, validator := range validatorsRes.Validators {
		anyPub, err := codectypes.NewAnyWithValue(validator.PubKey)
		if err != nil {
			return nil, err
		}
		outputValidatorsRes.Validators[i] = &Validator{
			Address:          validator.Address.String(),
			ProposerPriority: validator.ProposerPriority,
			PubKey:           anyPub,
			VotingPower:      validator.VotingPower,
		}
	}
	return outputValidatorsRes, nil
}

func (m *GetLatestValidatorSetResponse) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
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
func (s queryServer) GetValidatorSetByHeight(ctx context.Context, req *GetValidatorSetByHeightRequest) (*GetValidatorSetByHeightResponse, error) {
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

	outputValidatorsRes := &GetValidatorSetByHeightResponse{
		BlockHeight: validatorsRes.BlockHeight,
		Validators:  make([]*Validator, len(validatorsRes.Validators)),
	}

	for i, validator := range validatorsRes.Validators {
		anyPub, err := codectypes.NewAnyWithValue(validator.PubKey)
		if err != nil {
			return nil, err
		}
		outputValidatorsRes.Validators[i] = &Validator{
			Address:          validator.Address.String(),
			ProposerPriority: validator.ProposerPriority,
			PubKey:           anyPub,
			VotingPower:      validator.VotingPower,
		}
	}
	return outputValidatorsRes, nil
}

// GetNodeInfo implements ServiceServer.GetNodeInfo
func (s queryServer) GetNodeInfo(ctx context.Context, req *GetNodeInfoRequest) (*GetNodeInfoResponse, error) {
	status, err := getNodeStatus(s.clientCtx)
	if err != nil {
		return nil, err
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

// RegisterTendermintService registers the tendermint queries on the gRPC router.
func RegisterTendermintService(
	qrt gogogrpc.Server,
	clientCtx client.Context,
	interfaceRegistry codectypes.InterfaceRegistry,
) {
	RegisterServiceServer(
		qrt,
		NewQueryServer(clientCtx, interfaceRegistry),
	)
}

// RegisterGRPCGatewayRoutes mounts the tendermint service's GRPC-gateway routes on the
// given Mux.
func RegisterGRPCGatewayRoutes(clientConn gogogrpc.ClientConn, mux *runtime.ServeMux) {
	RegisterServiceHandlerClient(context.Background(), mux, NewServiceClient(clientConn))
}
