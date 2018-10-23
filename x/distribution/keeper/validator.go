package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// check whether a validator has distribution info
func (k Keeper) HasValidatorDistInfo(ctx sdk.Context,
	operatorAddr sdk.ValAddress) (exists bool) {
	store := ctx.KVStore(k.storeKey)
	return store.Has(GetValidatorDistInfoKey(operatorAddr))
}

// get the validator distribution info
func (k Keeper) GetValidatorDistInfo(ctx sdk.Context,
	operatorAddr sdk.ValAddress) (vdi types.ValidatorDistInfo) {

	store := ctx.KVStore(k.storeKey)

	b := store.Get(GetValidatorDistInfoKey(operatorAddr))
	if b == nil {
		panic("Stored validator-distribution info should not have been nil")
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

// remove a validator distribution info
func (k Keeper) RemoveValidatorDistInfo(ctx sdk.Context, valAddr sdk.ValAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(GetValidatorDistInfoKey(valAddr))
}

// withdrawal all the validator rewards including the commission
func (k Keeper) WithdrawValidatorRewardsAll(ctx sdk.Context, operatorAddr sdk.ValAddress) sdk.Error {

	if !k.HasValidatorDistInfo(ctx, operatorAddr) {
		return types.ErrNoValidatorDistInfo(k.codespace)
	}

	// withdraw self-delegation
	height := ctx.BlockHeight()
	validator := k.stakeKeeper.Validator(ctx, operatorAddr)
	lastValPower := k.stakeKeeper.GetLastValidatorPower(ctx, operatorAddr)
	accAddr := sdk.AccAddress(operatorAddr.Bytes())
	withdraw := k.getDelegatorRewardsAll(ctx, accAddr, height)

	// withdrawal validator commission rewards
	lastTotalPower := k.stakeKeeper.GetLastTotalPower(ctx)
	valInfo := k.GetValidatorDistInfo(ctx, operatorAddr)
	feePool := k.GetFeePool(ctx)
	valInfo, feePool, commission := valInfo.WithdrawCommission(feePool, height, lastTotalPower,
		lastValPower, validator.GetCommission())
	withdraw = withdraw.Plus(commission)
	k.SetValidatorDistInfo(ctx, valInfo)

	withdrawAddr := k.GetDelegatorWithdrawAddr(ctx, accAddr)
	truncated, change := withdraw.TruncateDecimal()
	feePool.CommunityPool = feePool.CommunityPool.Plus(change)
	k.SetFeePool(ctx, feePool)
	_, _, err := k.bankKeeper.AddCoins(ctx, withdrawAddr, truncated)
	if err != nil {
		panic(err)
	}

	return nil
}

// iterate over all the validator distribution infos (inefficient, just used to check invariants)
func (k Keeper) IterateValidatorDistInfos(ctx sdk.Context, fn func(index int64, distInfo types.ValidatorDistInfo) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, ValidatorDistInfoKey)
	defer iter.Close()
	index := int64(0)
	for ; iter.Valid(); iter.Next() {
		var vdi types.ValidatorDistInfo
		k.cdc.MustUnmarshalBinary(iter.Value(), &vdi)
		if fn(index, vdi) {
			return
		}
		index++
	}
}
