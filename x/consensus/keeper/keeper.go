package keeper

import (
	"context"
	"errors"
	"fmt"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttypes "github.com/cometbft/cometbft/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/event"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/consensus/exported"
	"github.com/cosmos/cosmos-sdk/x/consensus/types"
)

var StoreKey = "Consensus"

type Keeper struct {
	environment appmodule.Environment

	authority   string
	ParamsStore collections.Item[cmtproto.ConsensusParams]
	// storage of the last comet info
	cometInfo collections.Item[types.CometInfo]
}

var _ exported.ConsensusParamSetter = Keeper{}.ParamsStore

func NewKeeper(cdc codec.BinaryCodec, env appmodule.Environment, authority string) Keeper {
	sb := collections.NewSchemaBuilder(env.KVStoreService)
	return Keeper{
		environment: env,
		authority:   authority,
		ParamsStore: collections.NewItem(sb, collections.NewPrefix("Consensus"), "params", codec.CollValue[cmtproto.ConsensusParams](cdc)),
	}
}

func (k *Keeper) GetAuthority() string {
	return k.authority
}

// Querier

var _ types.QueryServer = Keeper{}

// Params queries params of consensus module
func (k Keeper) Params(ctx context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	params, err := k.ParamsStore.Get(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryParamsResponse{Params: &params}, nil
}

// MsgServer

var _ types.MsgServer = Keeper{}

func (k Keeper) UpdateParams(ctx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if k.GetAuthority() != msg.Authority {
		return nil, fmt.Errorf("invalid authority; expected %s, got %s", k.GetAuthority(), msg.Authority)
	}

	consensusParams, err := msg.ToProtoConsensusParams()
	if err != nil {
		return nil, err
	}
	if err := cmttypes.ConsensusParamsFromProto(consensusParams).ValidateBasic(); err != nil {
		return nil, err
	}

	if err := k.ParamsStore.Set(ctx, consensusParams); err != nil {
		return nil, err
	}

	if err := k.environment.EventService.EventManager(ctx).EmitKV(
		"update_consensus_params",
		event.NewAttribute("authority", msg.Authority),
		event.NewAttribute("parameters", consensusParams.String())); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}

// SetParams sets the consensus parameters on init of a chain. This is a consensus message. It can only be called by the consensus server
// This is used in the consensus message handler set in module.go.
func (k Keeper) SetParams(ctx context.Context, req *types.ConsensusMsgParams) (*types.ConsensusMsgParamsResponse, error) {
	consensusParams, err := req.ToProtoConsensusParams()
	if err != nil {
		return nil, err
	}
	if err := cmttypes.ConsensusParamsFromProto(consensusParams).ValidateBasic(); err != nil {
		return nil, err
	}

	if err := k.ParamsStore.Set(ctx, consensusParams); err != nil {
		return nil, err
	}

	return &types.ConsensusMsgParamsResponse{}, nil
}

func (k Keeper) GetCometInfo(ctx context.Context, req *types.MsgCometInfoRequest) (*types.MsgCometInfoResponse, error) {
	ci, err := k.cometInfo.Get(ctx)
	if errors.Is(err, collections.ErrNotFound) {
		ci := sdk.UnwrapSDKContext(ctx).CometInfo()
		res := &types.MsgCometInfoResponse{CometInfo: &types.CometInfo{
			ValidatorsHash:  ci.ValidatorsHash,
			ProposerAddress: ci.ProposerAddress,
			LastCommit:      ci.LastCommit, // TODO
			Evidence:        ci.Evidence,   // TODO
		}}

		return res, err
	}

	return &types.MsgCometInfoResponse{CometInfo: &ci}, err
}

func (k *Keeper) SetCometInfo(ctx context.Context, req *types.ConsensusMsgCometInfoRequest) (*types.ConsensusMsgCometInfoResponse, error) {
	err := k.cometInfo.Set(ctx, *req.CometInfo)

	return &types.ConsensusMsgCometInfoResponse{}, err
}
