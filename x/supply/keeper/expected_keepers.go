package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

// AccountKeeper defines the expected account keeper
type AccountKeeper interface {
	IterateAccounts(ctx sdk.Context, process func(auth.Account) (stop bool))
}

// CrisisKeeper defines the expected crisis keeper
type CrisisKeeper interface {
	RegisterRoute(moduleName, route string, invar sdk.Invariant)
}

// DistributionKeeper defines the expected distribution keeper
type DistributionKeeper interface {
	GetFeePoolCommunityCoins(ctx sdk.Context) sdk.DecCoins
	GetTotalRewards(ctx sdk.Context) sdk.DecCoins
}

// FeeCollectionKeeper defines the expected fee collection keeper
type FeeCollectionKeeper interface {
	GetCollectedFees(ctx sdk.Context) sdk.Coins
}

// StakingKeeper defines the expected staking keeper
type StakingKeeper interface {
	BondDenom(ctx sdk.Context) string
	StakingTokenSupply(ctx sdk.Context) sdk.Int
}
