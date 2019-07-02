package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
)

// AccountKeeper defines the expected account keeper (noalias)
type AccountKeeper interface {
	IterateAccounts(ctx sdk.Context, process func(exported.Account) (stop bool))
	GetAccount(sdk.Context, sdk.AccAddress) exported.Account
	SetAccount(sdk.Context, exported.Account)
	NewAccount(sdk.Context, exported.Account) exported.Account
}

// BankKeeper defines the expected bank keeper (noalias)
type BankKeeper interface {
	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) sdk.Error
	DelegateCoins(ctx sdk.Context, fromAdd, toAddr sdk.AccAddress, amt sdk.Coins) sdk.Error
	UndelegateCoins(ctx sdk.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) sdk.Error

	SubtractCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) (sdk.Coins, sdk.Error)
	AddCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) (sdk.Coins, sdk.Error)
}
