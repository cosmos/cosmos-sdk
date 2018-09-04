package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/stake"
	wire "github.com/tendermint/go-wire"
)

// keeper of the stake store
type Keeper struct {
	storeKey    sdk.StoreKey
	cdc         *wire.Codec
	coinKeeper  bank.Keeper
	stakeKeeper stake.Keeper

	// codespace
	codespace sdk.CodespaceType
}

func NewKeeper(cdc *wire.Codec, key sdk.StoreKey, ck bank.Keeper,
	sk stake.Keeper, codespace sdk.CodespaceType) Keeper {

	keeper := Keeper{
		storeKey:   key,
		cdc:        cdc,
		coinKeeper: ck,
		codespace:  codespace,
	}
	return keeper
}

//______________________________________________________________________

// get the global distribution info
func (k Keeper) GetGlobal(ctx sdk.Context) (global types.Global) {
	store := ctx.KVStore(k.storeKey)

	b := store.Get(GlobalKey)
	if b == nil {
		panic("Stored global should not have been nil")
	}

	k.cdc.MustUnmarshalBinary(b, &global)
	return
}

// set the global distribution info
func (k Keeper) SetGlobal(ctx sdk.Context, global types.Global) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshalBinary(global)
	store.Set(GlobalKey, b)
}

//______________________________________________________________________

// get the delegator distribution info
func (k Keeper) GetDelegatorDistInfo(ctx sdk.Context, delAddr sdk.AccAddress,
	valOperatorAddr sdk.ValAddress) (ddi types.DelegatorDistInfo) {

	store := ctx.KVStore(k.storeKey)

	b := store.Get(GetDelegationDistInfoKey(delAddr, valOperatorAddr))
	if b == nil {
		panic("Stored delegation-distribution info should not have been nil")
	}

	k.cdc.MustUnmarshalBinary(b, &ddi)
	return
}

// set the delegator distribution info
func (k Keeper) SetDelegatorDistInfo(ctx sdk.Context, ddi types.DelegatorDistInfo) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshalBinary(ddi)
	store.Set(GetDelegationDistInfoKey(ddi.DelegatorAddr, ddi.ValOperatorAddr), b)
}

//______________________________________________________________________

// get the validator distribution info
func (k Keeper) GetValidatorDistInfo(ctx sdk.Context,
	operatorAddr sdk.ValAddress) (vdi types.ValidatorDistInfo) {

	store := ctx.KVStore(k.storeKey)

	b := store.Get(GetValidatorDistInfoKey(operatorAddr))
	if b == nil {
		panic("Stored delegation-distribution info should not have been nil")
	}

	k.cdc.MustUnmarshalBinary(b, &vdi)
	return
}

// set the validator distribution info
func (k Keeper) SetValidatorDistInfo(ctx sdk.Context, vdi types.ValidatorDistInfo) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshalBinary(vdi)
	store.Set(GetValidatorDistInfoKey(vdi.OperatorAddr), b)
}
