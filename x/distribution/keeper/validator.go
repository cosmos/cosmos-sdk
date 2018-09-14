package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// get the validator distribution info
func (k Keeper) GetValidatorDistInfo(ctx sdk.Context,
	operatorAddr sdk.ValAddress) (vdi types.ValidatorDistInfo) {

	store := ctx.KVStore(k.storeKey)

	b := store.Get(GetValidatorDistInfoKey(ctx, operatorAddr))
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
	store.Set(GetValidatorDistInfoKey(ctx, vdi.OperatorAddr), b)
}

// XXX TODO
func (k Keeper) WithdrawValidatorRewardsAll(ctx sdk.Context,
	operatorAddr sdk.ValAddress, withdrawAddr sdk.AccAddress) {

	// withdraw self-delegation
	height := ctx.BlockHeight()
	validator := k.GetValidator(ctx, operatorAddr)
	withdraw := k.GetDelegatorRewardsAll(ctx, validator.OperatorAddr, height)

	// withdrawal validator commission rewards
	pool := k.stakeKeeper.GetPool(ctx)
	valInfo := k.GetValidatorDistInfo(ctx, operatorAddr)
	feePool := k.GetFeePool(ctx)
	feePool, commission := valInfo.WithdrawCommission(feePool, valInfo, height, pool.BondedTokens,
		validator.Tokens, validator.Commission)
	withdraw = withdraw.Add(commission)
	k.SetFeePool(feePool)

	k.coinKeeper.AddCoins(withdrawAddr, withdraw.TruncateDecimal())
}
