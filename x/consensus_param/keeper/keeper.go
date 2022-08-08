package keeper

import (
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/consensus_param/exported"
	"github.com/cosmos/cosmos-sdk/x/consensus_param/types"
)

type Keeper struct {
	storeKey    storetypes.StoreKey
	cdc         codec.BinaryCodec
	paramSetter exported.ConsensusParamSetter
	authority   string
}

func NewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey, paramSetter exported.ConsensusParamSetter, authority string) Keeper {
	return Keeper{
		storeKey:    storeKey,
		cdc:         cdc,
		paramSetter: paramSetter,
		authority:   authority,
	}
}

func (k *Keeper) SetParamSetter(paramSetter exported.ConsensusParamSetter) {
	k.paramSetter = paramSetter
}

func (k *Keeper) GetParamSetter() exported.ConsensusParamSetter {
	return k.paramSetter
}

func (k *Keeper) GetAuthority() string {
	return k.authority
}

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

	return store.Has(types.ParamStoreKeyConsensusParams)
}

func (k *Keeper) Set(ctx sdk.Context, cp *tmproto.ConsensusParams) {
	store := ctx.KVStore(k.storeKey)

	store.Set(types.ParamStoreKeyConsensusParams, k.cdc.MustMarshal(cp))
}
