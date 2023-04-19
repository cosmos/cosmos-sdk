package quarantine

import sdk "github.com/cosmos/cosmos-sdk/types"

var quarantineBypassKey = "bypass-quarantine-restriction"

// WithQuarantineBypass returns a new context that will cause the quarantine bank send restriction to be skipped.
func WithQuarantineBypass(ctx sdk.Context) sdk.Context {
	return ctx.WithValue(quarantineBypassKey, true)
}

// WithoutQuarantineBypass returns a new context that will cause the quarantine bank send restriction to not be skipped.
func WithoutQuarantineBypass(ctx sdk.Context) sdk.Context {
	return ctx.WithValue(quarantineBypassKey, false)
}

// HasQuarantineBypass checks the context to see if the quarantine bank send restriction should be skipped.
func HasQuarantineBypass(ctx sdk.Context) bool {
	bypassValue := ctx.Value(quarantineBypassKey)
	if bypassValue == nil {
		return false
	}
	bypass, isBool := bypassValue.(bool)
	return isBool && bypass
}
