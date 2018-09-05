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

// XXX TODO
func (k Keeper) WithdrawValidatorRewardsAll(ctx sdk.Context, operatorAddr, withdrawAddr sdk.AccAddress) {
	height = ctx.BlockHeight()
	feePool = k.GetFeePool(ctx)
	pool = k.stakeKeeper.GetPool(ctx)
	ValInfo = k.GetValidatorDistInfo(delegation.ValidatorAddr)
	validator = k.GetValidator(delegation.ValidatorAddr)

	// withdraw self-delegation
	withdraw = k.GetDelegatorRewardsAll(validator.OperatorAddr, height)

	// withdrawal validator commission rewards
	feePool, commission = valInfo.WithdrawCommission(feePool, valInfo, height, pool.BondedTokens,
		validator.Tokens, validator.Commission)
	withdraw += commission
	SetFeePool(feePool)

	k.coinKeeper.AddCoins(withdrawAddr, withdraw.TruncateDecimal())
}
