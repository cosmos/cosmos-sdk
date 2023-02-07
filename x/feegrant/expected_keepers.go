package feegrant

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AccountKeeper defines the expected auth Account Keeper (noalias)
type AccountKeeper interface {
	GetModuleAddress(moduleName string) sdk.AccAddress
	GetModuleAccount(ctx sdk.Context, moduleName string) sdk.ModuleAccountI

	NewAccountWithAddress(ctx sdk.Context, addr sdk.AccAddress) sdk.AccountI
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) sdk.AccountI
	SetAccount(ctx sdk.Context, acc sdk.AccountI)
}

// BankKeeper defines the expected supply Keeper (noalias)
type BankKeeper interface {
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
}
