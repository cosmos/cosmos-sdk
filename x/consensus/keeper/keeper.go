package keeper

import (
	"context"
	"errors"
	"fmt"

	cmtproto "github.com/cometbft/cometbft/api/cometbft/types/v1"
	cmttypes "github.com/cometbft/cometbft/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/appmodule"
	corecontext "cosmossdk.io/core/context"
	"cosmossdk.io/core/event"
	"cosmossdk.io/x/consensus/exported"
	"cosmossdk.io/x/consensus/types"

	"github.com/cosmos/cosmos-sdk/codec"
)

type Keeper struct {
	appmodule.Environment

	authority   string
	ParamsStore collections.Item[cmtproto.ConsensusParams]
}

var _ exported.ConsensusParamSetter = Keeper{}.ParamsStore

func NewKeeper(cdc codec.BinaryCodec, env appmodule.Environment, authority string) Keeper {
	sb := collections.NewSchemaBuilder(env.KVStoreService)
	return Keeper{
		Environment: env,
		authority:   authority,
		ParamsStore: collections.NewItem(sb, collections.NewPrefix("Consensus"), "params", codec.CollValue[cmtproto.ConsensusParams](cdc)),
	}
}

func (k *Keeper) GetAuthority() string {
	return k.authority
}

// InitGenesis initializes the initial state of the module
func (k *Keeper) InitGenesis(ctx context.Context) error {
	value, ok := ctx.Value(corecontext.InitInfoKey).(*types.MsgUpdateParams)
	if !ok {
		// no error for appv1 and appv2
		return nil
	}
	if value == nil {
		// no error for appv1
		return nil
	}

	consensusParams, err := value.ToProtoConsensusParams()
	if err != nil {
		return err
	}

	nextParams, err := k.paramCheck(ctx, consensusParams)
	if err != nil {
		return err
	}

	return k.ParamsStore.Set(ctx, nextParams.ToProto())
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

	nextParams, err := k.paramCheck(ctx, consensusParams)
	if err != nil {
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

// paramCheck validates the consensus params
func (k Keeper) paramCheck(ctx context.Context, consensusParams cmtproto.ConsensusParams) (*cmttypes.ConsensusParams, error) {
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

	return &nextParams, nil
}
