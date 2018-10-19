package mint

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	stake "github.com/cosmos/cosmos-sdk/x/stake"
)

// expected stake keeper
type StakeKeeper interface {
	GetPool(ctx sdk.Context) stake.Pool
	SetPool(ctx sdk.Context, pool stake.Pool)
	TotalPower(ctx sdk.Context) sdk.Dec
	BondedRatio(ctx sdk.Context) sdk.Dec
	InflateSupply(ctx sdk.Context, newTokens sdk.Dec)
}

// expected fee collection keeper interface
type FeeCollectionKeeper interface {
	AddCollectedFees(sdk.Context, sdk.Coins) sdk.Coins
}
