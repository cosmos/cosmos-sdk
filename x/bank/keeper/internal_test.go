package keeper

import "github.com/cosmos/cosmos-sdk/x/bank/types"

// UnsafeSetHooks updates the x/bank keeper's hooks, overriding any potential
// pre-existing hooks.
//
// WARNING: this function should only be used in tests.
func UnsafeSetHooks(k *BaseKeeper, h types.BankHooks) {
	k.hooks = h
}
