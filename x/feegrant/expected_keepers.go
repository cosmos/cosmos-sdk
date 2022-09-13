package feegrant

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	auth "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// AccountKeeper defines the expected auth Account Keeper (noalias)
type AccountKeeper interface {
	GetModuleAddress(moduleName string) sdk.AccAddress
	GetModuleAccount(ctx sdk.Context, moduleName string) auth.ModuleIAccount

	NewAccountWithAddress(ctx sdk.Context, addr sdk.AccAddress) auth.IAccount
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) auth.IAccount
	SetAccount(ctx sdk.Context, acc auth.IAccount)
}

// BankKeeper defines the expected supply Keeper (noalias)
type BankKeeper interface {
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
}
