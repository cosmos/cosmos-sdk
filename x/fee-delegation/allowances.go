package fee_delegation

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

type BasicFeeAllowance struct {
	// SpendLimit specifies the maximum amount of tokens that can be spent
	// by this capability and will be updated as tokens are spent. If it is
	// empty, there is no spend limit and any amount of coins can be spent.
	SpendLimit sdk.Coins
}

var _ FeeAllowance = BasicFeeAllowance{}

func (cap BasicFeeAllowance) Accept(fee sdk.Coins, block abci.Header) (allow bool, updated FeeAllowance, delete bool) {
	left, invalid := cap.SpendLimit.SafeSub(fee)
	if invalid {
		return false, nil, false
	}
	if left.IsZero() {
		return true, nil, true
	}
	return true, BasicFeeAllowance{SpendLimit: left}, false
}

