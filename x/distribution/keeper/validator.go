package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

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
func (k Keeper) WithdrawValidatorRewardsAll(ctx sdk.Context, operatorAddr sdk.ValAddress) {

	// withdraw self-delegation
	height := ctx.BlockHeight()
	validator := k.stakeKeeper.Validator(ctx, operatorAddr)
	accAddr := sdk.AccAddress(operatorAddr.Bytes())
	withdraw := k.getDelegatorRewardsAll(ctx, accAddr, height)

	// withdrawal validator commission rewards
	bondedTokens := k.stakeKeeper.TotalPower(ctx)
	valInfo := k.GetValidatorDistInfo(ctx, operatorAddr)
	feePool := k.GetFeePool(ctx)
	valInfo, feePool, commission := valInfo.WithdrawCommission(feePool, height, bondedTokens,
		validator.GetTokens(), validator.GetCommission())
	withdraw = withdraw.Plus(commission)
	k.SetValidatorDistInfo(ctx, valInfo)
	k.SetFeePool(ctx, feePool)

	withdrawAddr := k.GetDelegatorWithdrawAddr(ctx, accAddr)
	_, _, err := k.bankKeeper.AddCoins(ctx, withdrawAddr, withdraw.TruncateDecimal())
	if err != nil {
		panic(err)
	}
}
