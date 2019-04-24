package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SupplyKeeper defines the expected supply keeper
type SupplyKeeper interface {
	InflateSupply(ctx sdk.Context, supplyType string, amt sdk.Coins)
}

// StakingKeeper defines the expected staking keeper
type StakingKeeper interface {
	BondedRatio(ctx sdk.Context) sdk.Dec
	BondDenom(ctx sdk.Context) string
	InflateNotBondedTokenSupply(ctx sdk.Context, amt sdk.Int)
	StakingTokenSupply(ctx sdk.Context) sdk.Int
}

// FeeCollectionKeeper defines the expected fee collection keeper
type FeeCollectionKeeper interface {
	AddCollectedFees(sdk.Context, sdk.Coins) sdk.Coins
}
