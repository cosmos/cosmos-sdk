package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// Get the set of all validator-distribution-info's with no limits, used during genesis dump
func (k Keeper) GetAllValidatorDistInfos(ctx sdk.Context) (vdis []types.ValidatorDistInfo) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, ValidatorDistInfoKey)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var vdi types.ValidatorDistInfo
		k.cdc.MustUnmarshalBinaryLengthPrefixed(iterator.Value(), &vdi)
		vdis = append(vdis, vdi)
	}
	return vdis
}

// Get the set of all delegator-distribution-info's with no limits, used during genesis dump
func (k Keeper) GetAllDelegationDistInfos(ctx sdk.Context) (ddis []types.DelegationDistInfo) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, DelegationDistInfoKey)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var ddi types.DelegationDistInfo
		k.cdc.MustUnmarshalBinaryLengthPrefixed(iterator.Value(), &ddi)
		ddis = append(ddis, ddi)
	}
	return ddis
}

// Get the set of all delegator-withdraw addresses with no limits, used during genesis dump
func (k Keeper) GetAllDelegatorWithdrawInfos(ctx sdk.Context) (dwis []types.DelegatorWithdrawInfo) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, DelegatorWithdrawInfoKey)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		dw := types.DelegatorWithdrawInfo{
			DelegatorAddr: GetDelegatorWithdrawInfoAddress(iterator.Key()),
			WithdrawAddr:  sdk.AccAddress(iterator.Value()),
		}
		dwis = append(dwis, dw)
	}
	return dwis
}
