package crisis

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// expected bank keeper
type distrKeeper interface {
	DistributeFeePool(ctx sdk.Context, amount sdk.Coins, receiveAddr sdk.AccAddress) sdk.Error
}

// expected fee collection keeper
type feeCollectionKeeper interface {
	AddCollectedFees(ctx sdk.Context, coins sdk.Coins) sdk.Coins
}

// expected bank keeper
type bankKeeper interface {
	SubtractCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) (sdk.Coins, sdk.Error)
}
