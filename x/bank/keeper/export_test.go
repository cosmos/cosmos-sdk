package keeper

import "github.com/cosmos/cosmos-sdk/x/bank/types"

// This file exists in the keeper package to expose some private things
// for the purpose of testing in the keeper_test package.

// SetSendRestriction is a TEST ONLY method for overwriting the SendRestrictionFn.
func (k BaseSendKeeper) SetSendRestriction(restriction types.SendRestrictionFn) {
	k.sendRestriction.fn = restriction
}

// GetSendRestrictionFn is a TEST ONLY exposure of the currently defined SendRestrictionFn.
func (k BaseSendKeeper) GetSendRestrictionFn() types.SendRestrictionFn {
	return k.sendRestriction.fn
}

// SetLockedCoinsGetter is a TEST ONLY method for overwriting the GetLockedCoinsFn.
func (k BaseSendKeeper) SetLockedCoinsGetter(getter types.GetLockedCoinsFn) {
	k.lockedCoinsGetter.fn = getter
}

// GetLockedCoinsGetter is a TEST ONLY exposure of the currently defined GetLockedCoinsFn.
func (k BaseViewKeeper) GetLockedCoinsGetter() types.GetLockedCoinsFn {
	return k.lockedCoinsGetter.fn
}

// GetLockedCoinsFnWrapper is a TEST ONLY exposure of the getLockedCoinsFnWrapper function.
var GetLockedCoinsFnWrapper = getLockedCoinsFnWrapper
