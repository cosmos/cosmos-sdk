package keeper

import (
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/consensus/exported"
	"github.com/cosmos/cosmos-sdk/x/consensus/types"
)

var _ exported.ConsensusParamSetter = (*Keeper)(nil)

type Keeper struct {
	storeKey storetypes.StoreKey
	cdc      codec.BinaryCodec

	authority string
}

func NewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey, authority string) Keeper {
	return Keeper{
		storeKey:  storeKey,
		cdc:       cdc,
		authority: authority,
	}
}

func (k *Keeper) GetAuthority() string {
	return k.authority
}

// Get gets the consensus parameters
func (k *Keeper) Get(ctx sdk.Context) (*tmproto.ConsensusParams, error) {
	store := ctx.KVStore(k.storeKey)

	cp := &tmproto.ConsensusParams{}
	bz := store.Get(types.ParamStoreKeyConsensusParams)

	if err := k.cdc.Unmarshal(bz, cp); err != nil {
		return nil, err
	}

	return cp, nil
}

func (k *Keeper) Has(ctx sdk.Context) bool {
	store := ctx.KVStore(k.storeKey)

<<<<<<< HEAD
	return store.Has(types.ParamStoreKeyConsensusParams)
}

// Set sets the consensus parameters
func (k *Keeper) Set(ctx sdk.Context, cp *tmproto.ConsensusParams) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.ParamStoreKeyConsensusParams, k.cdc.MustMarshal(cp))
=======
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

	if err := k.event.EventManager(ctx).EmitKV(
		ctx,
		"update_consensus_params",
		event.Attribute{Key: "authority", Value: msg.Authority},
		event.Attribute{Key: "parameters", Value: consensusParams.String()}); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
>>>>>>> ed14ec03b (chore: check for nil params (#18041))
}
