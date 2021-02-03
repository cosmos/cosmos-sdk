package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authn "github.com/cosmos/cosmos-sdk/x/authn/types"
)

// AccountKeeper defines the expected auth Account Keeper (noalias)
type AccountKeeper interface {
	GetModuleAddress(moduleName string) sdk.AccAddress
	GetModuleAccount(ctx sdk.Context, moduleName string) authn.ModuleAccountI

	NewAccountWithAddress(ctx sdk.Context, addr sdk.AccAddress) authn.AccountI
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) authn.AccountI
	SetAccount(ctx sdk.Context, acc authn.AccountI)
}

// BankKeeper defines the expected supply Keeper (noalias)
type BankKeeper interface {
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
}
