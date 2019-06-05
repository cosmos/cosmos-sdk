package uniswap

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BankKeeper defines the expected coin keeper
type BankKeeper interface {
	AddCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) (sdk.Coins, sdk.Error)
}
