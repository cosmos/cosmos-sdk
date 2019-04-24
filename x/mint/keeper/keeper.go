package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/params"
)

// Keeper defines the keeper of the minting store
type Keeper struct {
	storeKey     sdk.StoreKey
	cdc          *codec.Codec
	paramSpace   params.Subspace
	supplyKeeper SupplyKeeper
	sk           StakingKeeper
	fck          FeeCollectionKeeper
}

// NewKeeper creates a new mint Keeper instance
func NewKeeper(cdc *codec.Codec, key sdk.StoreKey,
	paramSpace params.Subspace, supplyKeeper SupplyKeeper, sk StakingKeeper, fck FeeCollectionKeeper) Keeper {

	keeper := Keeper{
		storeKey:     key,
		cdc:          cdc,
		paramSpace:   paramSpace.WithKeyTable(ParamKeyTable()),
		supplyKeeper: supplyKeeper,
		sk:           sk,
		fck:          fck,
	}
	return keeper
}

// get the minter
func (k Keeper) GetMinter(ctx sdk.Context) (minter types.Minter) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(minterKey)
	if b == nil {
		panic("Stored minter should not have been nil")
	}
	k.cdc.MustUnmarshalBinaryLengthPrefixed(b, &minter)
	return
}

// set the minter
func (k Keeper) SetMinter(ctx sdk.Context, minter types.Minter) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshalBinaryLengthPrefixed(minter)
	store.Set(minterKey, b)
}
