package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type BankKeeper interface {
	AddCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) (sdk.Coins, sdk.Error)

	SetTotalSupply(ctx sdk.Context, totalSupply sdk.Coin)

	GetTotalSupply(ctx sdk.Context, denom string) (coin sdk.Coin, found bool)

	IncreaseTotalSupply(ctx sdk.Context, amt sdk.Coin) sdk.Error

	BurnCoins(ctx sdk.Context, fromAddr sdk.AccAddress, amt sdk.Coins) sdk.Error

	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) sdk.Error
}
