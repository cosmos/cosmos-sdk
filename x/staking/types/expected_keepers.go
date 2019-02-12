package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// expected coin keeper
type DistributionKeeper interface {
	GetFeePoolCommunityCoins(ctx sdk.Context) sdk.DecCoins
	GetOutstandingRewardsCoins(ctx sdk.Context) sdk.DecCoins
}

// expected fee collection keeper
type FeeCollectionKeeper interface {
	GetCollectedFees(ctx sdk.Context) sdk.Coins
}

// expected bank keeper
type BankKeeper interface {
	DelegateCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) (sdk.Tags, sdk.Error)
	UndelegateCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) (sdk.Tags, sdk.Error)
}
