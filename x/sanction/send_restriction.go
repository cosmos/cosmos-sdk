package sanction

import sdk "github.com/cosmos/cosmos-sdk/types"

var sanctionBypassKey = "bypass-sanction-restriction"

// WithSanctionBypass returns a new context that will cause the sanction bank send restriction to be skipped.
func WithSanctionBypass(ctx sdk.Context) sdk.Context {
	return ctx.WithValue(sanctionBypassKey, true)
}

// WithoutSanctionBypass returns a new context that will cause the sanction bank send restriction to not be skipped.
func WithoutSanctionBypass(ctx sdk.Context) sdk.Context {
	return ctx.WithValue(sanctionBypassKey, false)
}

// HasSanctionBypass checks the context to see if the sanction bank send restriction should be skipped.
func HasSanctionBypass(ctx sdk.Context) bool {
	bypassValue := ctx.Value(sanctionBypassKey)
	if bypassValue == nil {
		return false
	}
	bypass, isBool := bypassValue.(bool)
	return isBool && bypass
}
