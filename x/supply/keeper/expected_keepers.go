package keeper

import sdk "github.com/cosmos/cosmos-sdk/types"

// CrisisKeeper defines the expected crisis keeper
type CrisisKeeper interface {
	RegisterRoute(moduleName, route string, invar sdk.Invariant)
}

// DistributionKeeper defines the expected distribution keeper
type DistributionKeeper interface {
	GetFeePoolCommunityCoins(ctx sdk.Context) sdk.DecCoins
}

// FeeCollectionKeeper defines the expected fee collection keeper
type FeeCollectionKeeper interface {
	GetCollectedFees(ctx sdk.Context) sdk.Coins
}

// StakingKeeper defines the expected staking keeper
type StakingKeeper interface {
	BondDenom(ctx sdk.Context) string
	TotalBondedTokens(ctx sdk.Context) sdk.Int
}
