package mint

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
)

// BankKeeper defines the expected bank keeper
type BankKeeper interface {
	GetSupplier(ctx sdk.Context) bank.Supplier
	InflateSupply(ctx sdk.Context, amt sdk.Coins)
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
