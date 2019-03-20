package keeper

import (
	"container/list"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

const aminoCacheSize = 500

// keeper of the staking store
type Keeper struct {
	storeKey           sdk.StoreKey
	storeTKey          sdk.StoreKey
	cdc                *codec.Codec
	bankKeeper         types.BankKeeper
	hooks              sdk.StakingHooks
	paramstore         params.Subspace
	validatorCache     map[string]cachedValidator
	validatorCacheList *list.List

	// codespace
	codespace sdk.CodespaceType
}

func NewKeeper(cdc *codec.Codec, key, tkey sdk.StoreKey, bk types.BankKeeper,
	paramstore params.Subspace, codespace sdk.CodespaceType) Keeper {

	keeper := Keeper{
		storeKey:           key,
		storeTKey:          tkey,
		cdc:                cdc,
		bankKeeper:         bk,
		paramstore:         paramstore.WithKeyTable(ParamKeyTable()),
		hooks:              nil,
		validatorCache:     make(map[string]cachedValidator, aminoCacheSize),
		validatorCacheList: list.New(),
		codespace:          codespace,
	}
	return keeper
}

// Set the validator hooks
func (k *Keeper) SetHooks(sh sdk.StakingHooks) *Keeper {
	if k.hooks != nil {
		panic("cannot set validator hooks twice")
	}
	k.hooks = sh
	return k
}

// return the codespace
func (k Keeper) Codespace() sdk.CodespaceType {
	return k.codespace
}

// get the pool
func (k Keeper) GetPool(ctx sdk.Context) (pool types.Pool) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(PoolKey)
	if b == nil {
		panic("stored pool should not have been nil")
	}
	k.cdc.MustUnmarshalBinaryLengthPrefixed(b, &pool)
	return
}

// set the pool
func (k Keeper) SetPool(ctx sdk.Context, pool types.Pool) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshalBinaryLengthPrefixed(pool)
	store.Set(PoolKey, b)
}

// Load the last total validator power.
func (k Keeper) GetLastTotalPower(ctx sdk.Context) (power sdk.Int) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(LastTotalPowerKey)
	if b == nil {
		return sdk.ZeroInt()
	}
	k.cdc.MustUnmarshalBinaryLengthPrefixed(b, &power)
	return
}

// Set the last total validator power.
func (k Keeper) SetLastTotalPower(ctx sdk.Context, power sdk.Int) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshalBinaryLengthPrefixed(power)
	store.Set(LastTotalPowerKey, b)
}
