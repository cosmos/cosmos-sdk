package keeper

import (
	"errors"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// SetGovernanceDelegation sets a governance delegation in the store
func (keeper Keeper) SetGovernanceDelegation(ctx sdk.Context, delegation v1.GovernanceDelegation) {
	delAddr := sdk.MustAccAddressFromBech32(delegation.DelegatorAddress)
	if err := keeper.GovernanceDelegations.Set(ctx, delAddr, delegation); err != nil {
		panic(err)
	}

	// Set the reverse mapping from governor to delegation
	// mainly for querying all delegations for a governor
	govAddr := types.MustGovernorAddressFromBech32(delegation.GovernorAddress)
	if err := keeper.GovernanceDelegationsByGovernor.Set(ctx, collections.Join(govAddr, delAddr), delegation); err != nil {
		panic(err)
	}
}

// RemoveGovernanceDelegation removes a governance delegation from the store
func (keeper Keeper) RemoveGovernanceDelegation(ctx sdk.Context, delegatorAddr sdk.AccAddress) {
	// need to remove from both the delegator and governor mapping
	delegation, err := keeper.GovernanceDelegations.Get(ctx, delegatorAddr)
	if err != nil && !errors.Is(err, collections.ErrNotFound) {
		panic(err)
	}
	if errors.Is(err, collections.ErrNotFound) {
		return
	}
	govAddr := types.MustGovernorAddressFromBech32(delegation.GovernorAddress)
	keeper.GovernanceDelegations.Remove(ctx, delegatorAddr)
	keeper.GovernanceDelegationsByGovernor.Remove(ctx, collections.Join(govAddr, delegatorAddr))
}

// IncreaseGovernorShares increases the governor validator shares in the store
func (keeper Keeper) IncreaseGovernorShares(ctx sdk.Context, governorAddr types.GovernorAddress, validatorAddr sdk.ValAddress, shares math.LegacyDec) {
	valShares, err := keeper.ValidatorSharesByGovernor.Get(ctx, collections.Join(governorAddr, validatorAddr))
	if err != nil && !errors.Is(err, collections.ErrNotFound) {
		panic(err)
	}
	if errors.Is(err, collections.ErrNotFound) {
		valShares = v1.NewGovernorValShares(governorAddr, validatorAddr, shares)
	} else {
		valShares.Shares = valShares.Shares.Add(shares)
	}
	if err := keeper.ValidatorSharesByGovernor.Set(ctx, collections.Join(governorAddr, validatorAddr), valShares); err != nil {
		panic(err)
	}
}

// DecreaseGovernorShares decreases the governor validator shares in the store
func (keeper Keeper) DecreaseGovernorShares(ctx sdk.Context, governorAddr types.GovernorAddress, validatorAddr sdk.ValAddress, shares math.LegacyDec) {
	share, err := keeper.ValidatorSharesByGovernor.Get(ctx, collections.Join(governorAddr, validatorAddr))
	if err != nil && !errors.Is(err, collections.ErrNotFound) {
		panic(err)
	}
	if errors.Is(err, collections.ErrNotFound) {
		panic("cannot decrease shares for a non-existent governor delegation")
	}
	share.Shares = share.Shares.Sub(shares)
	if share.Shares.IsNegative() {
		panic("negative shares")
	}
	if share.Shares.IsZero() {
		keeper.ValidatorSharesByGovernor.Remove(ctx, collections.Join(governorAddr, validatorAddr))
	} else {
		if err := keeper.ValidatorSharesByGovernor.Set(ctx, collections.Join(governorAddr, validatorAddr), share); err != nil {
			panic(err)
		}
	}
}

// UndelegateFromGovernor decreases all governor validator shares in the store
// and then removes the governor delegation for the given delegator
func (keeper Keeper) UndelegateFromGovernor(ctx sdk.Context, delegatorAddr sdk.AccAddress) error {
	delegation, err := keeper.GovernanceDelegations.Get(ctx, delegatorAddr)
	if err != nil && !errors.Is(err, collections.ErrNotFound) {
		panic(err)
	}
	if errors.Is(err, collections.ErrNotFound) {
		return types.ErrGovernanceDelegationNotFound.Wrapf("governance delegation for delegator %s does not exist", delegatorAddr.String())
	}
	govAddr := types.MustGovernorAddressFromBech32(delegation.GovernorAddress)
	// iterate all delegations of delegator and decrease shares
	err = keeper.sk.IterateDelegations(ctx, delegatorAddr, func(_ int64, delegation stakingtypes.DelegationI) (stop bool) {
		valAddr, err := sdk.ValAddressFromBech32(delegation.GetValidatorAddr())
		if err != nil {
			panic(err) // This should never happen
		}
		keeper.DecreaseGovernorShares(ctx, govAddr, valAddr, delegation.GetShares())
		return false
	})
	if err != nil {
		return sdkerrors.ErrInvalidRequest.Wrapf("failed to iterate delegations: %v", err)
	}
	// remove the governor delegation
	keeper.RemoveGovernanceDelegation(ctx, delegatorAddr)
	return nil
}

// DelegateGovernor creates a governor delegation for the given delegator
// and increases all governor validator shares in the store
func (keeper Keeper) DelegateToGovernor(ctx sdk.Context, delegatorAddr sdk.AccAddress, governorAddr types.GovernorAddress) error {
	delegation := v1.NewGovernanceDelegation(delegatorAddr, governorAddr)
	keeper.SetGovernanceDelegation(ctx, delegation)
	// iterate all delegations of delegator and increase shares
	err := keeper.sk.IterateDelegations(ctx, delegatorAddr, func(_ int64, delegation stakingtypes.DelegationI) (stop bool) {
		valAddr, err := sdk.ValAddressFromBech32(delegation.GetValidatorAddr())
		if err != nil {
			panic(err) // This should never happen
		}
		keeper.IncreaseGovernorShares(ctx, governorAddr, valAddr, delegation.GetShares())
		return false
	})
	if err != nil {
		return sdkerrors.ErrInvalidRequest.Wrapf("failed to iterate delegations: %v", err)
	}
	return nil
}

// RedelegateGovernor re-delegates all governor validator shares from one governor to another
func (keeper Keeper) RedelegateToGovernor(ctx sdk.Context, delegatorAddr sdk.AccAddress, dstGovernorAddr types.GovernorAddress) error {
	// undelegate from the source governor
	if err := keeper.UndelegateFromGovernor(ctx, delegatorAddr); err != nil {
		return err
	}
	// delegate to the destination governor
	return keeper.DelegateToGovernor(ctx, delegatorAddr, dstGovernorAddr)
}
