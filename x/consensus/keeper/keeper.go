package keeper

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	cmtproto "github.com/cometbft/cometbft/api/cometbft/types/v1"
	cmttypes "github.com/cometbft/cometbft/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"cosmossdk.io/collections"
	coreapp "cosmossdk.io/core/app"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/event"
	"cosmossdk.io/x/consensus/exported"
	"cosmossdk.io/x/consensus/types"

	"github.com/cosmos/cosmos-sdk/codec"
)

var StoreKey = "Consensus"

type Keeper struct {
	appmodule.Environment

	authority   string
	ParamsStore collections.Item[cmtproto.ConsensusParams]
	// storage of the last comet info
	cometInfo collections.Item[types.CometInfo]
}

var _ exported.ConsensusParamSetter = Keeper{}.ParamsStore

func NewKeeper(cdc codec.BinaryCodec, env appmodule.Environment, authority string) Keeper {
	sb := collections.NewSchemaBuilder(env.KVStoreService)
	return Keeper{
		Environment: env,
		authority:   authority,
		ParamsStore: collections.NewItem(sb, collections.NewPrefix("Consensus"), "params", codec.CollValue[cmtproto.ConsensusParams](cdc)),
		cometInfo:   collections.NewItem(sb, collections.NewPrefix("CometInfo"), "comet_info", codec.CollValue[types.CometInfo](cdc)),
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

	paramsProto, err := k.ParamsStore.Get(ctx)

	var params cmttypes.ConsensusParams
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			params = cmttypes.ConsensusParams{}
		} else {
			return nil, err
		}
	} else {
		params = cmttypes.ConsensusParamsFromProto(paramsProto)
	}

	nextParams := params.Update(&consensusParams)

	if err := nextParams.ValidateBasic(); err != nil {
		return nil, err
	}

	if err := params.ValidateUpdate(&consensusParams, k.HeaderService.HeaderInfo(ctx).Height); err != nil {
		return nil, err
	}

	if err := k.ParamsStore.Set(ctx, nextParams.ToProto()); err != nil {
		return nil, err
	}

	if err := k.EventService.EventManager(ctx).EmitKV(
		"update_consensus_params",
		event.NewAttribute("authority", msg.Authority),
		event.NewAttribute("parameters", consensusParams.String())); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}

func (k Keeper) SetCometInfo(ctx context.Context, msg *types.MsgSetCometInfo) (*types.MsgSetCometInfoResponse, error) {
	if !bytes.Equal(coreapp.ConsensusIdentity, []byte(msg.Authority)) {
		return nil, fmt.Errorf("invalid authority; expected %s, got %s", coreapp.ConsensusIdentity, msg.Authority)
	}

	cometInfo := types.CometInfo{
		Evidence:        msg.Evidence,
		ValidatorsHash:  msg.ValidatorsHash,
		ProposerAddress: msg.ProposerAddress,
		LastCommit:      msg.LastCommit,
	}

	if err := k.cometInfo.Set(ctx, cometInfo); err != nil {
		return nil, err
	}

	return &types.MsgSetCometInfoResponse{}, nil
}

func (k Keeper) GetCometInfo(ctx context.Context, _ *types.QueryGetCometInfoRequest) (*types.QueryGetCometInfoResponse, error) {
	cometInfo, err := k.cometInfo.Get(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryGetCometInfoResponse{CometInfo: &cometInfo}, nil
}
