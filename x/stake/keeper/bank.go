package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// load a delegator bond
func (k PrivlegedKeeper) AddCoins(ctx sdk.Context, amount int64, address sdk.Address) {
	denom := k.GetParams(ctx).BondDenom
	coins := sdk.Coins{{denom, amount}}
	k.coinKeeper.AddCoins(ctx, address, coins)
}
