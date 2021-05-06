package keeper

import "github.com/cosmos/cosmos-sdk/x/gov/types"

// UnsafeSetHooks updates the gov keepers hooks, overriding any potential
// pre-existing hooks.
// WARNING: this function should only be used in tests.
func UnsafeSetHooks(k *Keeper, h types.GovHooks) {
	k.hooks = h
}
