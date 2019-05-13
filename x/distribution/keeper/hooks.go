package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Wrapper struct
type Hooks struct {
	k Keeper
}

var _ sdk.StakingHooks = Hooks{}

// Create new distribution hooks
func (k Keeper) Hooks() Hooks { return Hooks{k} }

// nolint
func (h Hooks) AfterValidatorCreated(ctx sdk.Context, valAddr sdk.ValAddress) {
	val := h.k.stakingKeeper.Validator(ctx, valAddr)
	h.k.initializeValidator(ctx, val)
}
func (h Hooks) BeforeValidatorModified(ctx sdk.Context, valAddr sdk.ValAddress) {
}
func (h Hooks) AfterValidatorRemoved(ctx sdk.Context, _ sdk.ConsAddress, valAddr sdk.ValAddress) {

	// fetch outstanding
	outstanding := h.k.GetValidatorOutstandingRewards(ctx, valAddr)

	// force-withdraw commission
	commission := h.k.GetValidatorAccumulatedCommission(ctx, valAddr)
	if !commission.IsZero() {
		// subtract from outstanding
		outstanding = outstanding.Sub(commission)

		// split into integral & remainder
		coins, remainder := commission.TruncateDecimal()

		// remainder to community pool
		feePool := h.k.GetFeePool(ctx)
		feePool.CommunityPool = feePool.CommunityPool.Add(remainder)
		h.k.SetFeePool(ctx, feePool)

		// add to validator account
		if !coins.IsZero() {

			accAddr := sdk.AccAddress(valAddr)
			withdrawAddr := h.k.GetDelegatorWithdrawAddr(ctx, accAddr)

			if _, err := h.k.bankKeeper.AddCoins(ctx, withdrawAddr, coins); err != nil {
				panic(err)
			}
		}
	}

	// add outstanding to community pool
	feePool := h.k.GetFeePool(ctx)
	feePool.CommunityPool = feePool.CommunityPool.Add(outstanding)
	h.k.SetFeePool(ctx, feePool)

	// delete outstanding
	h.k.DeleteValidatorOutstandingRewards(ctx, valAddr)

	// remove commission record
	h.k.DeleteValidatorAccumulatedCommission(ctx, valAddr)

	// clear slashes
	h.k.DeleteValidatorSlashEvents(ctx, valAddr)

	// clear historical rewards
	h.k.DeleteValidatorHistoricalRewards(ctx, valAddr)

	// clear current rewards
	h.k.DeleteValidatorCurrentRewards(ctx, valAddr)
}
func (h Hooks) BeforeDelegationCreated(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
	val := h.k.stakingKeeper.Validator(ctx, valAddr)

	// increment period
	h.k.incrementValidatorPeriod(ctx, val)
}
func (h Hooks) BeforeDelegationSharesModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
	val := h.k.stakingKeeper.Validator(ctx, valAddr)
	del := h.k.stakingKeeper.Delegation(ctx, delAddr, valAddr)

	// withdraw delegation rewards (which also increments period)
	if _, err := h.k.withdrawDelegationRewards(ctx, val, del); err != nil {
		panic(err)
	}
}
func (h Hooks) BeforeDelegationRemoved(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
	// nothing needed here since BeforeDelegationSharesModified will always also be called
}
func (h Hooks) AfterDelegationModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
	// create new delegation period record
	h.k.initializeDelegation(ctx, valAddr, delAddr)
}
func (h Hooks) AfterValidatorBeginUnbonding(ctx sdk.Context, _ sdk.ConsAddress, valAddr sdk.ValAddress) {
}
func (h Hooks) AfterValidatorBonded(ctx sdk.Context, _ sdk.ConsAddress, valAddr sdk.ValAddress) {
}
func (h Hooks) BeforeValidatorSlashed(ctx sdk.Context, valAddr sdk.ValAddress, fraction sdk.Dec) {
	// record the slash event
	h.k.updateValidatorSlashFraction(ctx, valAddr, fraction)
}
