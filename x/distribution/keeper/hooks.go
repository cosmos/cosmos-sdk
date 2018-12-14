package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

func (k Keeper) onValidatorCreated(ctx sdk.Context, valAddr sdk.ValAddress) {
	// Set current rewards
	k.setValidatorCurrentRewards(ctx, valAddr, types.ValidatorCurrentRewards{})

	// Set accumulated commission
	k.setValidatorAccumulatedCommission(ctx, valAddr, types.ValidatorAccumulatedCommission{})

	// Create a new fees / power record
}

func (k Keeper) onValidatorStakeModified(ctx sdk.Context, valAddr sdk.ValAddress) {
	// Make a new fees / power record
}

func (k Keeper) onDelegationCreated(ctx sdk.Context, delAddr sdk.AccAddress,
	valAddr sdk.ValAddress) {
	// Create a new fees / power record
}

// Withdrawal all validator rewards
func (k Keeper) onDelegationSharesModified(ctx sdk.Context, delAddr sdk.AccAddress,
	valAddr sdk.ValAddress) {
	// Withdraw old delegation
	// Create new delegation
}

// Withdrawal all validator distribution rewards and cleanup the distribution record
func (k Keeper) onDelegationRemoved(ctx sdk.Context, delAddr sdk.AccAddress,
	valAddr sdk.ValAddress) {
	// Withdraw rewards
}

// Wrapper struct
type Hooks struct {
	k Keeper
}

var _ sdk.StakingHooks = Hooks{}

// Create new distribution hooks
func (k Keeper) Hooks() Hooks { return Hooks{k} }

// nolint
func (h Hooks) OnValidatorCreated(ctx sdk.Context, valAddr sdk.ValAddress) {
	h.k.onValidatorCreated(ctx, valAddr)
}
func (h Hooks) OnValidatorModified(ctx sdk.Context, valAddr sdk.ValAddress) {
}
func (h Hooks) OnValidatorRemoved(ctx sdk.Context, _ sdk.ConsAddress, valAddr sdk.ValAddress) {
}
func (h Hooks) OnDelegationCreated(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
}
func (h Hooks) OnDelegationSharesModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
}
func (h Hooks) OnDelegationRemoved(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
}
func (h Hooks) OnValidatorBeginUnbonding(ctx sdk.Context, _ sdk.ConsAddress, valAddr sdk.ValAddress) {
}
func (h Hooks) OnValidatorBonded(ctx sdk.Context, _ sdk.ConsAddress, valAddr sdk.ValAddress) {
}
func (h Hooks) OnValidatorPowerDidChange(ctx sdk.Context, _ sdk.ConsAddress, valAddr sdk.ValAddress) {
}
