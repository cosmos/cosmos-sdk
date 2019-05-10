package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/supply"
)

// expected coin keeper
type DistributionKeeper interface {
	GetFeePoolCommunityCoins(ctx sdk.Context) sdk.DecCoins
	GetValidatorOutstandingRewardsCoins(ctx sdk.Context, val sdk.ValAddress) sdk.DecCoins
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

// expected crisis keeper
type CrisisKeeper interface {
	RegisterRoute(moduleName, route string, invar sdk.Invariant)
}

// SupplyKeeper defines the expected supply Keeper
type SupplyKeeper interface {
	Deflate(ctx sdk.Context, amount sdk.Coins)

	GetPoolAccountByName(ctx sdk.Context, name string) (supply.PoolAccount, sdk.Error)
	SetPoolAccount(ctx sdk.Context, macc supply.PoolAccount)

	GetCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	SendCoinsPoolToPool(ctx sdk.Context, senderModule, recipientModule string, amt sdk.Coins) sdk.Error
	BurnCoins(ctx sdk.Context, name string, amt sdk.Coins) sdk.Error
}
