package keeper

import "github.com/cosmos/cosmos-sdk/x/bank/types"

// This file exists in the keeper package to expose some private things
// for the purpose of testing in the keeper_test package.

func (k BaseSendKeeper) SetSendRestriction(restriction types.SendRestrictionFn) {
	k.sendRestriction.fn = restriction
}

func (k BaseSendKeeper) GetSendRestrictionFn() types.SendRestrictionFn {
	return k.sendRestriction.fn
}
