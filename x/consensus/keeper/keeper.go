package keeper

import (
	"context"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttypes "github.com/cometbft/cometbft/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/event"
	storetypes "cosmossdk.io/core/store"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/consensus/exported"
	"github.com/cosmos/cosmos-sdk/x/consensus/types"
)

var StoreKey = "Consensus"

type Keeper struct {
	storeService storetypes.KVStoreService
	event        event.Service

	authority   string
	ParamsStore collections.Item[cmtproto.ConsensusParams]
}

var _ exported.ConsensusParamSetter = Keeper{}.ParamsStore

func NewKeeper(cdc codec.BinaryCodec, storeService storetypes.KVStoreService, authority string, em event.Service) Keeper {
	sb := collections.NewSchemaBuilder(storeService)
	return Keeper{
		storeService: storeService,
		authority:    authority,
		event:        em,
		ParamsStore:  collections.NewItem(sb, collections.NewPrefix("Consensus"), "params", codec.CollValue[cmtproto.ConsensusParams](cdc)),
	}
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

func (k *Keeper) GetAuthority() string {
	return k.authority
}

// UpdateParams updates consensus parameters. Note that the new authority value
// takes effect at the start of the next block, when BeginBlock loads fresh
// consensus params from the store. Within the same block as this update,
// ValidateAuthority still checks against the old authority.
func (k Keeper) UpdateParams(ctx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if err := sdkCtx.ValidateAuthority(k.authority, msg.Authority); err != nil {
		return nil, err
	}

	consensusParams, err := msg.ToProtoConsensusParams()
	if err != nil {
		return nil, err
	}

	paramsProto, err := k.ParamsStore.Get(ctx)
	if err != nil {
		return nil, err
	}

	// initialize version params with zero value if not set
	if paramsProto.Version == nil {
		paramsProto.Version = &cmtproto.VersionParams{}
	}

	params := cmttypes.ConsensusParamsFromProto(paramsProto)

	nextParams := params.Update(&consensusParams)

	if err := nextParams.ValidateBasic(); err != nil {
		return nil, err
	}

	if err := params.ValidateUpdate(&consensusParams, sdkCtx.BlockHeader().Height); err != nil {
		return nil, err
	}

	if err := k.ParamsStore.Set(ctx, nextParams.ToProto()); err != nil {
		return nil, err
	}

	if err := k.event.EventManager(ctx).EmitKV(
		ctx,
		"update_consensus_params",
		event.Attribute{Key: "authority", Value: msg.Authority},
		event.Attribute{Key: "parameters", Value: consensusParams.String()}); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}
