package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/slashing/internal/types"
)

// Unjail calls the staking Unjail function to unjail a validator if the
// jailed period has concluded
func (k Keeper) Unjail(ctx sdk.Context, validatorAddr sdk.ValAddress) sdk.Error {
	validator := k.sk.Validator(ctx, validatorAddr)
	if validator == nil {
		return types.ErrNoValidatorForAddress(k.codespace)
	}

	// cannot be unjailed if no self-delegation exists
	selfDel := k.sk.Delegation(ctx, sdk.AccAddress(validatorAddr), validatorAddr)
	if selfDel == nil {
		return types.ErrMissingSelfDelegation(k.codespace)
	}

	if validator.TokensFromShares(selfDel.GetShares()).TruncateInt().LT(validator.GetMinSelfDelegation()) {
		return types.ErrSelfDelegationTooLowToUnjail(k.codespace)
	}

	// cannot be unjailed if not jailed
	if !validator.IsJailed() {
		return types.ErrValidatorNotJailed(k.codespace)
	}

	consAddr := sdk.ConsAddress(validator.GetConsPubKey().Address())

	info, found := k.GetValidatorSigningInfo(ctx, consAddr)
	if !found {
		return types.ErrNoValidatorForAddress(k.codespace)
	}

	// cannot be unjailed if tombstoned
	if info.Tombstoned {
		return types.ErrValidatorJailed(k.codespace)
	}

	// cannot be unjailed until out of jail
	if ctx.BlockHeader().Time.Before(info.JailedUntil) {
		return types.ErrValidatorJailed(k.codespace)
	}

	k.sk.Unjail(ctx, consAddr)
	return nil
}
