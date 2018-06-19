package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"

	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

// keeper of the stake store
type Keeper struct {
	storeKey   sdk.StoreKey
	cdc        *wire.Codec
	coinKeeper bank.Keeper

	// codespace
	codespace sdk.CodespaceType
}

func NewKeeper(cdc *wire.Codec, key sdk.StoreKey, ck bank.Keeper, codespace sdk.CodespaceType) Keeper {
	keeper := Keeper{
		storeKey:   key,
		cdc:        cdc,
		coinKeeper: ck,
		codespace:  codespace,
	}
	return keeper
}

//_________________________________________________________________________

// full permission keeper of the stake store
type PrivilegedKeeper struct {
	Keeper

	storeKey   sdk.StoreKey
	cdc        *wire.Codec
	coinKeeper bank.Keeper

	// codespace
	codespace sdk.CodespaceType
}

func NewPrivilegedKeeper(cdc *wire.Codec, key sdk.StoreKey, ck bank.Keeper, codespace sdk.CodespaceType) PrivilegedKeeper {
	keeper := PrivilegedKeeper{
		Keeper:     NewKeeper(cdc, key, ck, codespace),
		storeKey:   key,
		cdc:        cdc,
		coinKeeper: ck,
		codespace:  codespace,
	}
	return keeper
}

//_________________________________________________________________________

// return the codespace
func (k Keeper) Codespace() sdk.CodespaceType {
	return k.codespace
}

// return the codespace
func (k PrivilegedKeeper) CoinKeeper() bank.Keeper {
	return k.coinKeeper
}

//_________________________________________________________________________
// some generic reads/writes that don't need their own files

// load/save the global staking params
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	store := ctx.KVStore(k.storeKey)

	b := store.Get(ParamKey)
	if b == nil {
		panic("Stored params should not have been nil")
	}

	k.cdc.MustUnmarshalBinary(b, &params)
	return
}

// Need a distinct function because setParams depends on an existing previous
// record of params to exist (to check if maxValidators has changed) - and we
// panic on retrieval if it doesn't exist - hence if we use setParams for the very
// first params set it will panic.
func (k PrivilegedKeeper) SetNewParams(ctx sdk.Context, params types.Params) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshalBinary(params)
	store.Set(ParamKey, b)
}

// set the params
func (k PrivilegedKeeper) SetParams(ctx sdk.Context, params types.Params) {
	store := ctx.KVStore(k.storeKey)
	exParams := k.GetParams(ctx)

	// if max validator count changes, must recalculate validator set
	if exParams.MaxValidators != params.MaxValidators {
		k.UpdateBondedValidatorsFull(ctx)
	}
	b := k.cdc.MustMarshalBinary(params)
	store.Set(ParamKey, b)
}

//_______________________________________________________________________

// load/save the pool
func (k Keeper) GetPool(ctx sdk.Context) (pool types.Pool) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(PoolKey)
	if b == nil {
		panic("Stored pool should not have been nil")
	}
	k.cdc.MustUnmarshalBinary(b, &pool)
	return
}

// set the pool
func (k PrivilegedKeeper) SetPool(ctx sdk.Context, pool types.Pool) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshalBinary(pool)
	store.Set(PoolKey, b)
}

//__________________________________________________________________________

// get the current in-block validator operation counter
func (k Keeper) GetIntraTxCounter(ctx sdk.Context) int16 {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(IntraTxCounterKey)
	if b == nil {
		return 0
	}
	var counter int16
	k.cdc.MustUnmarshalBinary(b, &counter)
	return counter
}

// set the current in-block validator operation counter
func (k PrivilegedKeeper) SetIntraTxCounter(ctx sdk.Context, counter int16) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinary(counter)
	store.Set(IntraTxCounterKey, bz)
}
