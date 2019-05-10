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

type DistributionKeeper interface {
	GetValidatorOutstandingRewardsCoins(ctx sdk.Context, valAddr sdk.ValAddress) sdk.Coins
}

// StakingKeeper defines the expected staking keeper
type StakingKeeper interface {
	BondDenom(ctx sdk.Context) string
	StakingTokenSupply(ctx sdk.Context) sdk.Int
	IterateValidators(ctx sdk.Context, handler func(index int64, validator sdk.Validator) (stop bool))
}
