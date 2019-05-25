package ibc

import sdk "github.com/YunSuk-Yeo/cosmos-sdk/types"

// expected bank keeper
type BankKeeper interface {
	AddCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) (sdk.Coins, sdk.Tags, sdk.Error)
	SubtractCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) (sdk.Coins, sdk.Tags, sdk.Error)
}
