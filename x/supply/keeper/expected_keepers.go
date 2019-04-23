package keeper

import sdk "github.com/cosmos/cosmos-sdk/types"

// CrisisKeeper defunes the expected crisis keeper
type CrisisKeeper interface {
	RegisterRoute(moduleName, route string, invar sdk.Invariant)
}

// // FeeCollectionKeeper defines the expected fee collection keeper
// type FeeCollectionKeeper interface {
// 	AddCollectedFees(sdk.Context, sdk.Coins) sdk.Coins
// }

// StakingKeeper defines the expected staking keeper
type StakingKeeper interface {
	BondDenom(ctx sdk.Context) string
	TotalBondedTokens(ctx sdk.Context) sdk.Int
}
