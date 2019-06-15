package types // noalias

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/supply"
)

// StakingKeeper defines the expected staking keeper
type StakingKeeper interface {
	StakingTokenSupply(ctx sdk.Context) sdk.Int
	BondedRatio(ctx sdk.Context) sdk.Dec
}

// SupplyKeeper defines the expected supply keeper
type SupplyKeeper interface {
	GetModuleAccountByName(ctx sdk.Context, name string) supply.ModuleAccount
	SetModuleAccount(ctx sdk.Context, macc supply.ModuleAccount)

	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) sdk.Error
	SendCoinsFromModuleToModule(ctx sdk.Context, senderModule, recipientModule string, amt sdk.Coins) sdk.Error
	MintCoins(ctx sdk.Context, name string, amt sdk.Coins) sdk.Error
}
