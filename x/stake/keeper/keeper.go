package keeper

import (
	"container/list"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

const aminoCacheSize = 500

// keeper of the stake store
type Keeper struct {
	storeKey           sdk.StoreKey
	storeTKey          sdk.StoreKey
	cdc                *codec.Codec
	bankKeeper         bank.Keeper
	hooks              sdk.StakingHooks
	paramstore         params.Subspace
	validatorCache     map[string]cachedValidator
	validatorCacheList *list.List

	// codespace
	codespace sdk.CodespaceType
}

func NewKeeper(cdc *codec.Codec, key, tkey sdk.StoreKey, ck bank.Keeper, paramstore params.Subspace, codespace sdk.CodespaceType) Keeper {
	keeper := Keeper{
		storeKey:           key,
		storeTKey:          tkey,
		cdc:                cdc,
		bankKeeper:         ck,
		paramstore:         paramstore.WithTypeTable(ParamTypeTable()),
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

//_________________________________________________________________________

// return the codespace
func (k Keeper) Codespace() sdk.CodespaceType {
	return k.codespace
}

//_______________________________________________________________________

// load the pool
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

//_______________________________________________________________________

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

//_______________________________________________________________________

// Load the last validator power.
// Returns zero if the operator was not a validator last block.
func (k Keeper) GetLastValidatorPower(ctx sdk.Context, operator sdk.ValAddress) (power sdk.Int) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(GetLastValidatorPowerKey(operator))
	if bz == nil {
		return sdk.ZeroInt()
	}
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &power)
	return
}

// Set the last validator power.
func (k Keeper) SetLastValidatorPower(ctx sdk.Context, operator sdk.ValAddress, power sdk.Int) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(power)
	store.Set(GetLastValidatorPowerKey(operator), bz)
}

// Iterate over last validator powers.
func (k Keeper) IterateLastValidatorPowers(ctx sdk.Context, handler func(operator sdk.ValAddress, power sdk.Int) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, LastValidatorPowerKey)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		addr := sdk.ValAddress(iter.Key()[len(LastValidatorPowerKey):])
		var power sdk.Int
		k.cdc.MustUnmarshalBinaryLengthPrefixed(iter.Value(), &power)
		if handler(addr, power) {
			break
		}
	}
}

// Delete the last validator power.
func (k Keeper) DeleteLastValidatorPower(ctx sdk.Context, operator sdk.ValAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(GetLastValidatorPowerKey(operator))
}
