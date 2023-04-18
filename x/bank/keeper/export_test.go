package keeper

// This file exists in the keeper package to expose some private things
// for the purpose of testing in the keeper_test package.

func (k BaseSendKeeper) SetSendRestriction(restriction SendRestrictionFn) {
	k.sendRestriction.Fn = restriction
}

func (k BaseSendKeeper) GetSendRestrictionFn() SendRestrictionFn {
	return k.sendRestriction.Fn
}
