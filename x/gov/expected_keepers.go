package gov

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/supply"
)

// AccountKeeper defines the expected account keeper
type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) auth.Account
	SetAccount(ctx sdk.Context, acc auth.Account)
}

// BankKeeper defines the expected bank keeper
type BankKeeper interface {
	GetCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	SubtractCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) (sdk.Coins, sdk.Error)
	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) sdk.Error
}

// SupplyKeeper defines the expected supply keeper
type SupplyKeeper interface {
	InflateSupply(ctx sdk.Context, supplyType string, amount sdk.Coins)
	DeflateSupply(ctx sdk.Context, supplyType string, amount sdk.Coins)
	SetSupplier(ctx sdk.Context, supplier supply.Supplier) // used for testing
}
