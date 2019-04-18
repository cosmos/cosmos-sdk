package gov

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/supply"
)

// AccountKeeper defines the expected account keeper
type AccountKeeper interface {
	NewAccount() (ctx sdk.Context, acc auth.Account) auth.Account
	SetAccount(ctx sdk.Context, acc Account)
}

// BankKeeper defines the expected bank keeper
type BankKeeper interface {
	GetCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	AddCoins(tx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) (sdk.Coins, sdk.Error)
	SubtractCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) (sdk.Coins, sdk.Error)
}

// SupplyKeeper defines the expected supply keeper
type SupplyKeeper interface {
	InflateSupply(ctx sdk.Context, supplyType string, amount sdk.Coins)
}
