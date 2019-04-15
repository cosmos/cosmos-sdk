package gov

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
)

// expected bank keeper
type BankKeeper interface {
	GetCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	AddCoins(tx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) (sdk.Coins, sdk.Error)

	AddTokenHolder(ctx sdk.Context, moduleName string) (bank.TokenHolder, sdk.Error)
	RequestTokens(ctx sdk.Context, moduleName string, amount sdk.Coins) sdk.Error
	RelinquishTokens(ctx sdk.Context, moduleName string, amount sdk.Coins) sdk.Error
}
