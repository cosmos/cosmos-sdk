package types

import sdk "github.com/cosmos/cosmos-sdk/types"

const bypassKey = "bypass-vesting-locked-coins" //nolint:gosec // Not actually credentials.

// WithVestingLockedBypass returns a new context that will cause the vesting locked coins lookup to be skipped.
func WithVestingLockedBypass(ctx sdk.Context) sdk.Context {
	return ctx.WithValue(bypassKey, true)
}

// WithoutVestingLockedBypass returns a new context that will cause the vesting locked coins lookup to not be skipped.
func WithoutVestingLockedBypass(ctx sdk.Context) sdk.Context {
	return ctx.WithValue(bypassKey, false)
}

// HasVestingLockedBypass checks the context to see if the vesting locked coins lookup should be skipped.
func HasVestingLockedBypass(ctx sdk.Context) bool {
	bypassValue := ctx.Value(bypassKey)
	if bypassValue == nil {
		return false
	}
	bypass, isBool := bypassValue.(bool)
	return isBool && bypass
}
