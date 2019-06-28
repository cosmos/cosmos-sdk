package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	// TODO: uncomment when supply module is merged to master
	//supply "github.com/cosmos/cosmos-sdk/x/supply/exported"
)

// BankKeeper defines the expected bank keeper
type BankKeeper interface {
	HasCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) bool
	GetCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
}

// SupplyKeeper defines the expected supply keeper
type SupplyKeeper interface {
	//	GetModuleAccount(ctx sdk.Context, moduleName string) supply.ModuleAccountI

	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) sdk.Error
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) sdk.Error

	MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) sdk.Error
	BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) sdk.Error
}
