package keeper

import (
	context "context"
	"errors"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// Hooks wrapper struct for gov keeper
type Hooks struct {
	k Keeper
}

var _ stakingtypes.StakingHooks = Hooks{}

// Return the staking hooks
func (keeper Keeper) StakingHooks() Hooks {
	return Hooks{keeper}
}

// BeforeDelegationSharesModified is called when a delegation's shares are modified
// We trigger a governor shares decrease here subtracting all delegation shares.
// The right amount of shares will be possibly added back in AfterDelegationModified
func (h Hooks) BeforeDelegationSharesModified(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	// does the delegator have a governance delegation?
	govDelegation, err := h.k.GovernanceDelegations.Get(ctx, delAddr)
	if err != nil && !errors.Is(err, collections.ErrNotFound) {
		return err
	}
	if errors.Is(err, collections.ErrNotFound) {
		return nil
	}
	govAddr := types.MustGovernorAddressFromBech32(govDelegation.GovernorAddress)

	// Fetch the delegation
	delegation, _ := h.k.sk.GetDelegation(ctx, delAddr, valAddr)

	// update the Governor's Validator shares
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	h.k.DecreaseGovernorShares(sdkCtx, govAddr, valAddr, delegation.Shares)

	return nil
}

// AfterDelegationModified is called when a delegation is created or modified
// We trigger a governor shares increase here adding all delegation shares.
// It is balanced by the full-amount decrease in BeforeDelegationSharesModified
func (h Hooks) AfterDelegationModified(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	// does the delegator have a governance delegation?
	govDelegation, err := h.k.GovernanceDelegations.Get(ctx, delAddr)
	if err != nil && !errors.Is(err, collections.ErrNotFound) {
		return err
	}
	if errors.Is(err, collections.ErrNotFound) {
		return nil
	}

	// Fetch the delegation
	delegation, err := h.k.sk.GetDelegation(ctx, delAddr, valAddr)
	if err != nil {
		return err
	}

	govAddr := types.MustGovernorAddressFromBech32(govDelegation.GovernorAddress)

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	// Calculate the new shares and update the Governor's shares
	shares := delegation.Shares
	h.k.IncreaseGovernorShares(sdkCtx, govAddr, valAddr, shares)

	// if the delegator is also an active governor, ensure min self-delegation requirement is met,
	// otherwise set governor to inactive
	delGovAddr := types.GovernorAddress(delAddr.Bytes())
	if governor, err := h.k.Governors.Get(ctx, delGovAddr); err != nil && governor.IsActive() {
		if governor.GetAddress().String() != govDelegation.GovernorAddress {
			panic("active governor delegating to another governor")
		}
		// if the governor no longer meets the min self-delegation, set to inactive
		if !h.k.ValidateGovernorMinSelfDelegation(sdkCtx, governor) {
			governor.Status = v1.Inactive
			now := sdkCtx.BlockTime()
			governor.LastStatusChangeTime = &now
			if err := h.k.Governors.Set(ctx, governor.GetAddress(), governor); err != nil {
				return err
			}
		}
	}

	return nil
}

// BeforeDelegationRemoved is called when a delegation is removed
// We verify if the delegator is also an active governor and if so check
// that the min self-delegation requirement is still met, otherwise set governor
// status to inactive
func (h Hooks) BeforeDelegationRemoved(ctx context.Context, delAddr sdk.AccAddress, _ sdk.ValAddress) error {
	// if the delegator is also an active governor, ensure min self-delegation requirement is met,
	// otherwise set governor to inactive
	delGovAddr := types.GovernorAddress(delAddr.Bytes())
	if governor, err := h.k.Governors.Get(ctx, delGovAddr); err != nil && governor.IsActive() {
		govDelegation, err := h.k.GovernanceDelegations.Get(ctx, delAddr)
		if err != nil && !errors.Is(err, collections.ErrNotFound) {
			return err
		}
		if errors.Is(err, collections.ErrNotFound) {
			panic("active governor without governance self-delegation")
		}
		if governor.GetAddress().String() != govDelegation.GovernorAddress {
			panic("active governor delegating to another governor")
		}
		// if the governor no longer meets the min self-delegation, set to inactive
		sdkCtx := sdk.UnwrapSDKContext(ctx)
		if !h.k.ValidateGovernorMinSelfDelegation(sdkCtx, governor) {
			governor.Status = v1.Inactive
			now := sdkCtx.BlockTime()
			governor.LastStatusChangeTime = &now
			if err := h.k.Governors.Set(ctx, governor.GetAddress(), governor); err != nil {
				return err
			}
		}
	}

	return nil
}

func (h Hooks) BeforeValidatorSlashed(_ context.Context, _ sdk.ValAddress, _ math.LegacyDec) error {
	return nil
}

func (h Hooks) AfterValidatorCreated(_ context.Context, _ sdk.ValAddress) error {
	return nil
}

func (h Hooks) BeforeValidatorModified(_ context.Context, _ sdk.ValAddress) error {
	return nil
}

func (h Hooks) AfterValidatorRemoved(_ context.Context, _ sdk.ConsAddress, _ sdk.ValAddress) error {
	return nil
}

func (h Hooks) AfterValidatorBonded(_ context.Context, _ sdk.ConsAddress, _ sdk.ValAddress) error {
	return nil
}

func (h Hooks) AfterValidatorBeginUnbonding(_ context.Context, _ sdk.ConsAddress, _ sdk.ValAddress) error {
	return nil
}

func (h Hooks) BeforeDelegationCreated(_ context.Context, _ sdk.AccAddress, _ sdk.ValAddress) error {
	return nil
}

func (h Hooks) AfterUnbondingInitiated(_ context.Context, _ uint64) error {
	return nil
}
